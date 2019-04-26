// +build cgo
// +build gpu

package cl

import (
	"encoding/binary"
	"errors"
	"sync/atomic"

	. "github.com/iotaledger/iota.go/const"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
)

var loopcount byte = 32

var countCL int64

type bufferInfo struct {
	size    int64
	flag    cl.MemFlag
	isArray bool
	data    []byte
}

func init() {
	// TODO: update to Curl-P-81
	// pows["PowCL"] = PowCL
}

var stopCL = true

// nolint: gocyclo
func exec(
	que *cl.CommandQueue,
	ker []*cl.Kernel,
	cores, nlocal int,
	mobj []*cl.MemObject,
	founded *int32,
	tryte chan trinary.Trytes) error {

	// initialize
	nglobal := cores * nlocal
	ev1, err := que.EnqueueNDRangeKernel(ker[0], nil, []int{nglobal}, []int{nlocal}, nil)
	if err != nil {
		return err
	}
	defer ev1.Release()

	if err = que.Finish(); err != nil {
		return err
	}

	found := make([]byte, 1)
	var cnt int
	num := int64(cores) * 64 * int64(loopcount)
	for cnt = 0; found[0] == 0 && *founded == 0 && !stopCL; cnt++ {
		// start searching
		var ev2 *cl.Event
		ev2, err = que.EnqueueNDRangeKernel(ker[1], nil, []int{nglobal}, []int{nlocal}, nil)
		if err != nil {
			return err
		}
		defer ev2.Release()

		ev3, err := que.EnqueueReadBufferBytes(mobj[6], true, found, []*cl.Event{ev2})
		if err != nil {
			return err
		}
		ev3.Release()

		atomic.AddInt64(&countCL, num)
	}

	if *founded != 0 || stopCL {
		return nil
	}

	atomic.StoreInt32(founded, 1)

	// finalize, get the result.
	ev4, err := que.EnqueueNDRangeKernel(ker[2], nil, []int{nglobal}, []int{nlocal}, nil)
	if err != nil {
		return err
	}
	defer ev4.Release()

	result := make([]byte, curl.HashSize*8)
	ev5, err := que.EnqueueReadBufferBytes(mobj[0], true, result, []*cl.Event{ev4})
	if err != nil {
		return err
	}
	ev5.Release()

	rr := make(trinary.Trits, curl.HashSize)
	for i := 0; i < curl.HashSize; i++ {
		switch {
		case result[i*8] == 0xff:
			rr[i] = -1
		case result[i*8] == 0x0 && result[i*8+7] == 0x0:
			rr[i] = 0
		case result[i*8] == 0x1 || result[i*8+7] == 0x1:
			rr[i] = 1
		}
	}

	tryte <- rr.MustTrytes()
	return nil
}

// nolint: gocyclo
func loopCL(binfo []bufferInfo) (trinary.Trytes, error) {
	defers := make([]func(), 0, 10)
	defer func() {
		for _, f := range defers {
			f()
		}
	}()

	platforms, err := cl.GetPlatforms()
	if err != nil {
		return "", err
	}

	exist := false
	var founded int32
	result := make(chan trinary.Trytes)
	for _, p := range platforms {
		var devs []*cl.Device
		devs, err = p.GetDevices(cl.DeviceTypeGPU)
		if err != nil || len(devs) == 0 {
			continue
		}

		exist = true
		// TODO: this case checks the error after appending, but all the other cases below
		// do it the opposite way. Check to see if this can be reversed to maintain the
		// pattern
		cont, err := cl.CreateContext(devs)
		defers = append(defers, cont.Release)
		if err != nil {
			return "", err
		}

		prog, err := cont.CreateProgramWithSource([]string{kernel})
		if err != nil {
			return "", err
		}

		defers = append(defers, prog.Release)
		if err := prog.BuildProgram(devs, "-Werror"); err != nil {
			println(p.Name())
			return "", err
		}

		ker := make([]*cl.Kernel, 3)
		defers = append(defers, func() {
			for _, k := range ker {
				if k != nil {
					k.Release()
				}
			}
		})

		for i, n := range []string{"init", "search", "finalize"} {
			ker[i], err = prog.CreateKernel(n)
			if err != nil {
				return "", err
			}
		}

		for _, d := range devs {
			mult := d.MaxWorkGroupSize()
			cores := d.MaxComputeUnits()
			mmax := d.MaxMemAllocSize()
			isLittle := d.EndianLittle()
			nlocal := 0
			for nlocal = curl.StateSize; nlocal > mult; {
				nlocal /= 3
			}

			var totalmem int64
			mobj := make([]*cl.MemObject, len(binfo))
			defers = append(defers, func() {
				for _, o := range mobj {
					if o != nil {
						o.Release()
					}
				}
			})

			que, err := cont.CreateCommandQueue(d, 0)
			if err != nil {
				return "", err
			}

			defers = append(defers, que.Release)

			for i, inf := range binfo {
				msize := inf.size
				if inf.isArray {
					msize *= int64(cores * mult)
				}

				if totalmem += msize; totalmem > mmax {
					return "", errors.New("max memory passed")
				}

				mobj[i], err = cont.CreateEmptyBuffer(inf.flag, int(msize))
				if err != nil {
					return "", err
				}

				if inf.data != nil {
					var ev *cl.Event
					switch {
					case isLittle:
						ev, err = que.EnqueueWriteBufferBytes(mobj[i], true, inf.data, nil)
					default:
						data := make([]byte, len(inf.data))
						for i := range inf.data {
							data[i] = inf.data[len(inf.data)-i-1]
						}
						ev, err = que.EnqueueWriteBufferBytes(mobj[i], true, data, nil)
					}

					if err != nil {
						return "", err
					}
					ev.Release()
				}

				for _, k := range ker {
					if err := k.SetArg(i, mobj[i]); err != nil {
						return "", err
					}
				}
			}

			go func() {
				err := exec(que, ker, cores, nlocal, mobj, &founded, result)
				if err != nil {
					panic(err)
				}
			}()
		}
	}

	if !exist {
		return "", errors.New("no GPU found")
	}

	r := <-result
	close(result)

	return r, nil
}

// PowCL is proof of work of iota in OpenCL.
func PowCL(trytes trinary.Trytes, mwm int) (trinary.Trytes, error) {
	switch {
	case !stopCL:
		stopCL = true
		return "", errors.New("pow is already running, stopped")
	case trytes == "":
		return "", errors.New("invalid trytes")
	}

	stopCL = false
	countCL = 0

	tr := MustTrytesToTrits(trytes)

	c := curl.NewCurlP81().(*curl.Curl)
	c.Absorb(tr[:(TransactionTrinarySize - HashTrinarySize)])
	copy(c.State, tr[TransactionTrinarySize-HashTrinarySize:])

	lmid, hmid := para(c.State)
	lmid[0] = low0
	hmid[0] = high0
	lmid[1] = low1
	hmid[1] = high1
	lmid[2] = low2
	hmid[2] = high2
	lmid[3] = low3
	hmid[3] = high3

	low := make([]byte, 8*curl.StateSize)
	for i, v := range lmid {
		binary.LittleEndian.PutUint64(low[8*i:], v)
	}

	high := make([]byte, 8*curl.StateSize)
	for i, v := range hmid {
		binary.LittleEndian.PutUint64(high[8*i:], v)
	}

	binfo := []bufferInfo{
		bufferInfo{
			8 * curl.HashSize, cl.MemWriteOnly, false, nil,
		},
		bufferInfo{
			8 * curl.StateSize, cl.MemReadWrite, true, low, //mid_low
		},
		bufferInfo{
			8 * curl.StateSize, cl.MemReadWrite, true, high, //mid_high
		},
		bufferInfo{
			8 * curl.StateSize, cl.MemReadWrite, true, nil,
		},
		bufferInfo{
			8 * curl.StateSize, cl.MemReadWrite, true, nil,
		},
		bufferInfo{
			8, cl.MemWriteOnly, false, []byte{byte(mwm), 0, 0, 0, 0, 0, 0, 0}, // mwm
		},
		bufferInfo{
			1, cl.MemReadWrite, false, nil,
		},
		bufferInfo{
			8, cl.MemReadWrite, true, nil,
		},
		bufferInfo{
			8, cl.MemWriteOnly, false, []byte{loopcount, 0, 0, 0, 0, 0, 0, 0}, //loop_count
		},
	}

	return loopCL(binfo)
}
