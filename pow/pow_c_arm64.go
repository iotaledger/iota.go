// +build cgo
// +build pow_arm_c128
// +build linux
// +build arm64

package pow

// #cgo LDFLAGS:
// #cgo CFLAGS: -Wall
/*
#include <stdio.h>
#include <string.h>

#include "arm_neon.h"

#define HBITS 0xFFFFFFFFFFFFFFFFuLL
#define LBITS 0x0000000000000000uLL
#define HASH_LENGTH 243              //trits
#define NONCE_LENGTH 81              //trits
#define STATE_LENGTH 3 * HASH_LENGTH //trits
#define TX_LENGTH 2673               //trytes
#define INCR_START HASH_LENGTH - NONCE_LENGTH + 4 + 27

#define LOW00 0xDB6DB6DB6DB6DB6DuLL  //0b1101101101101101101101101101101101101101101101101101101101101101L;
#define HIGH00 0xB6DB6DB6DB6DB6DBuLL //0b1011011011011011011011011011011011011011011011011011011011011011L;
#define LOW10 0xF1F8FC7E3F1F8FC7uLL  //0b1111000111111000111111000111111000111111000111111000111111000111L;
#define HIGH10 0x8FC7E3F1F8FC7E3FuLL //0b1000111111000111111000111111000111111000111111000111111000111111L;
#define LOW20 0x7FFFE00FFFFC01FFuLL  //0b0111111111111111111000000000111111111111111111000000000111111111L;
#define HIGH20 0xFFC01FFFF803FFFFuLL //0b1111111111000000000111111111111111111000000000111111111111111111L;
#define LOW30 0xFFC0000007FFFFFFuLL  //0b1111111111000000000000000000000000000111111111111111111111111111L;
#define HIGH30 0x003FFFFFFFFFFFFFuLL //0b0000000000111111111111111111111111111111111111111111111111111111L;
#define LOW40 0xFFFFFFFFFFFFFFFFuLL  //0b1111111111111111111111111111111111111111111111111111111111111111L;
#define HIGH40 0xFFFFFFFFFFFFFFFFuLL //0b1111111111111111111111111111111111111111111111111111111111111111L;

#define LOW01 0x6DB6DB6DB6DB6DB6uLL  //0b0110110110110110110110110110110110110110110110110110110110110110
#define HIGH01 0xDB6DB6DB6DB6DB6DuLL //0b1101101101101101101101101101101101101101101101101101101101101101
#define LOW11 0xF8FC7E3F1F8FC7E3uLL  //0b1111100011111100011111100011111100011111100011111100011111100011
#define HIGH11 0xC7E3F1F8FC7E3F1FuLL //0b1100011111100011111100011111100011111100011111100011111100011111
#define LOW21 0xC01FFFF803FFFF00uLL  //0b1100000000011111111111111111100000000011111111111111111100000000
#define HIGH21 0x3FFFF007FFFE00FFuLL //0b0011111111111111111100000000011111111111111111100000000011111111
#define LOW31 0x00000FFFFFFFFFFFuLL  //0b0000000000000000000011111111111111111111111111111111111111111111
#define HIGH31 0xFFFFFFFFFFFE0000uLL //0b1111111111111111111111111111111111111111111111100000000000000000
#define LOW41 0x000000000001FFFFuLL  //0b0000000000000000000000000000000000000000000000011111111111111111
#define HIGH41 0xFFFFFFFFFFFFFFFFuLL //0b1111111111111111111111111111111111111111111111111111111111111111

const int indices[] = {
    0, 364, 728, 363, 727, 362, 726, 361, 725, 360, 724, 359, 723, 358, 722, 357, 721, 356, 720, 355, 719, 354, 718, 353, 717, 352, 716, 351, 715, 350, 714, 349, 713, 348, 712, 347, 711, 346, 710, 345, 709, 344, 708, 343, 707, 342, 706, 341, 705, 340, 704, 339, 703, 338, 702, 337, 701, 336, 700, 335, 699, 334, 698, 333, 697, 332, 696, 331, 695, 330, 694, 329, 693, 328, 692, 327, 691, 326, 690, 325, 689, 324, 688, 323, 687, 322, 686, 321, 685, 320, 684, 319, 683, 318, 682, 317, 681, 316, 680, 315, 679, 314, 678, 313, 677, 312, 676, 311, 675, 310, 674, 309, 673, 308, 672, 307, 671, 306, 670, 305, 669, 304, 668, 303, 667, 302, 666, 301, 665, 300, 664, 299, 663, 298, 662, 297, 661, 296, 660, 295, 659, 294, 658, 293, 657, 292, 656, 291, 655, 290, 654, 289, 653, 288, 652, 287, 651, 286, 650, 285, 649, 284, 648, 283, 647, 282, 646, 281, 645, 280, 644, 279, 643, 278, 642, 277, 641, 276, 640, 275, 639, 274, 638, 273, 637, 272, 636, 271, 635, 270, 634, 269, 633, 268, 632, 267, 631, 266, 630, 265, 629, 264, 628, 263, 627, 262, 626, 261, 625, 260, 624, 259, 623, 258, 622, 257, 621, 256, 620, 255, 619, 254, 618, 253, 617, 252, 616, 251, 615, 250, 614, 249, 613, 248, 612, 247, 611, 246, 610, 245, 609, 244, 608, 243, 607, 242, 606, 241, 605, 240, 604, 239, 603, 238, 602, 237, 601, 236, 600, 235, 599, 234, 598, 233, 597, 232, 596, 231, 595, 230, 594, 229, 593, 228, 592, 227, 591, 226, 590, 225, 589, 224, 588, 223, 587, 222, 586, 221, 585, 220, 584, 219, 583, 218, 582, 217, 581, 216, 580, 215, 579, 214, 578, 213, 577, 212, 576, 211, 575, 210, 574, 209, 573, 208, 572, 207, 571, 206, 570, 205, 569, 204, 568, 203, 567, 202, 566, 201, 565, 200, 564, 199, 563, 198, 562, 197, 561, 196, 560, 195, 559, 194, 558, 193, 557, 192, 556, 191, 555, 190, 554, 189, 553, 188, 552, 187, 551, 186, 550, 185, 549, 184, 548, 183, 547, 182, 546, 181, 545, 180, 544, 179, 543, 178, 542, 177, 541, 176, 540, 175, 539, 174, 538, 173, 537, 172, 536, 171, 535, 170, 534, 169, 533, 168, 532, 167, 531, 166, 530, 165, 529, 164, 528, 163, 527, 162, 526, 161, 525, 160, 524, 159, 523, 158, 522, 157, 521, 156, 520, 155, 519, 154, 518, 153, 517, 152, 516, 151, 515, 150, 514, 149, 513, 148, 512, 147, 511, 146, 510, 145, 509, 144, 508, 143, 507, 142, 506, 141, 505, 140, 504, 139, 503, 138, 502, 137, 501, 136, 500, 135, 499, 134, 498, 133, 497, 132, 496, 131, 495, 130, 494, 129, 493, 128, 492, 127, 491, 126, 490, 125, 489, 124, 488, 123, 487, 122, 486, 121, 485, 120, 484, 119, 483, 118, 482, 117, 481, 116, 480, 115, 479, 114, 478, 113, 477, 112, 476, 111, 475, 110, 474, 109, 473, 108, 472, 107, 471, 106, 470, 105, 469, 104, 468, 103, 467, 102, 466, 101, 465, 100, 464, 99, 463, 98, 462, 97, 461, 96, 460, 95, 459, 94, 458, 93, 457, 92, 456, 91, 455, 90, 454, 89, 453, 88, 452, 87, 451, 86, 450, 85, 449, 84, 448, 83, 447, 82, 446, 81, 445, 80, 444, 79, 443, 78, 442, 77, 441, 76, 440, 75, 439, 74, 438, 73, 437, 72, 436, 71, 435, 70, 434, 69, 433, 68, 432, 67, 431, 66, 430, 65, 429, 64, 428, 63, 427, 62, 426, 61, 425, 60, 424, 59, 423, 58, 422, 57, 421, 56, 420, 55, 419, 54, 418, 53, 417, 52, 416, 51, 415, 50, 414, 49, 413, 48, 412, 47, 411, 46, 410, 45, 409, 44, 408, 43, 407, 42, 406, 41, 405, 40, 404, 39, 403, 38, 402, 37, 401, 36, 400, 35, 399, 34, 398, 33, 397, 32, 396, 31, 395, 30, 394, 29, 393, 28, 392, 27, 391, 26, 390, 25, 389, 24, 388, 23, 387, 22, 386, 21, 385, 20, 384, 19, 383, 18, 382, 17, 381, 16, 380, 15, 379, 14, 378, 13, 377, 12, 376, 11, 375, 10, 374, 9, 373, 8, 372, 7, 371, 6, 370, 5, 369, 4, 368, 3, 367, 2, 366, 1, 365, 0};

void transformARM64(uint64x2_t *lmid, uint64x2_t *hmid)
{
  int j, r, t1, t2;
  uint64x2_t alpha, beta, gamma, delta;
  uint64x2_t *lto = lmid + STATE_LENGTH, *hto = hmid + STATE_LENGTH;
  uint64x2_t *lfrom = lmid, *hfrom = hmid;
  for (r = 0; r < 80; r++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      t1 = indices[j];
      t2 = indices[j + 1];

      alpha = lfrom[t1];
      beta = hfrom[t1];
      gamma = hfrom[t2];
      delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);

      lto[j] = ~delta;
      hto[j] = (alpha ^ gamma) | delta;
    }
    uint64x2_t *lswap = lfrom, *hswap = hfrom;
    lfrom = lto;
    hfrom = hto;
    lto = lswap;
    hto = hswap;
  }
  for (j = 0; j < HASH_LENGTH; j++)
  {
    t1 = indices[j];
    t2 = indices[j + 1];

    alpha = lfrom[t1];
    beta = hfrom[t1];
    gamma = hfrom[t2];
    delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);

    lto[j] = ~delta;
    hto[j] = (alpha ^ gamma) | delta;
  }
}

int incrARM64(uint64x2_t *mid_low, uint64x2_t *mid_high)
{
  int i;

  uint64x2_t carry = {LOW00, LOW01};

  for (i = INCR_START; i < HASH_LENGTH && (i == INCR_START || carry[0]); i++)
  {
    uint64x2_t low = mid_low[i], high = mid_high[i];
    mid_low[i] = high ^ low;
    mid_high[i] = low;
    carry = high & (~low);
  }
  return i == HASH_LENGTH;
}

void seriARM64(uint64x2_t *low, uint64x2_t *high, int n, signed char *r)
{
  int i = 0, index = 0;
  if (n > 63)
  {
    n -= 64;
    index = 1;
  }
  for (i = HASH_LENGTH-NONCE_LENGTH; i < HASH_LENGTH; i++)
  {
    unsigned long long ll = (low[i][index] >> n) & 1;
    unsigned long long hh = (high[i][index] >> n) & 1;
    if (hh == 0 && ll == 1)
    {
      r[i+NONCE_LENGTH-HASH_LENGTH] = -1;
    }
    if (hh == 1 && ll == 1)
    {
      r[i+NONCE_LENGTH-HASH_LENGTH] = 0;
    }
    if (hh == 1 && ll == 0)
    {
      r[i+NONCE_LENGTH-HASH_LENGTH] = 1;
    }
  }
}

int checkARM64(uint64x2_t *l, uint64x2_t *h, int m)
{
  int i, j; //omit init for speed

  uint64x2_t nonce_probe = {HBITS, HBITS};

  for (i = HASH_LENGTH - m; i < HASH_LENGTH; i++)
  {
    nonce_probe &= ~(l[i] ^ h[i]);
    if (nonce_probe[0] == LBITS && nonce_probe[1] == LBITS)
    {
      return -1;
    }
  }
  for (j = 0; j < 2; j++)
  {
    for (i = 0; i < 64; i++)
    {
      if ((nonce_probe[j] >> i) & 1)
      {
        return i + j * 64;
      }
    }
  }
  return -2;
}

long long int loopARM64(uint64x2_t *lmid, uint64x2_t *hmid, int m, signed char *nonce, int *stop)
{
  int n = 0, j = 0;
  long long int i = 0;

  uint64x2_t lcpy[STATE_LENGTH * 2], hcpy[STATE_LENGTH * 2];
  for (i = 0; !incrARM64(lmid, hmid) && !*stop; i++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      lcpy[j] = lmid[j];
      hcpy[j] = hmid[j];
    }
    transformARM64(lcpy, hcpy);
    if ((n = checkARM64(lcpy + STATE_LENGTH, hcpy + STATE_LENGTH, m)) >= 0)
    {
      seriARM64(lmid, hmid, n, nonce);
      return i * 128;
    }
  }
  return -i*128-1;
}

void paraARM64(signed char in[], uint64x2_t l[], uint64x2_t h[])
{
  int i = 0;
  for (i = 0; i < STATE_LENGTH; i++)
  {
    switch (in[i])
    {
    case 0:
      l[i][0] = HBITS;
      l[i][1] = HBITS;
      h[i][0] = HBITS;
      h[i][1] = HBITS;
      break;
    case 1:
      l[i][0] = LBITS;
      l[i][1] = LBITS;
      h[i][0] = HBITS;
      h[i][1] = HBITS;
      break;
    case -1:
      l[i][0] = HBITS;
      l[i][1] = HBITS;
      h[i][0] = LBITS;
      h[i][1] = LBITS;
      break;
    }
  }
}

void incrNARM64(int n, uint64x2_t *mid_low, uint64x2_t *mid_high)
{
  int i,j;
  for (j=0;j<n;j++){
    uint64x2_t carry = {HBITS, HBITS};

    for (i = HASH_LENGTH * 2/3 + 4; i < HASH_LENGTH * 2/3 + 4 + 27 &&  carry[0]; i++)
    {
      uint64x2_t low = mid_low[i], high = mid_high[i];
      mid_low[i] = high ^ low;
      mid_high[i] = low;
      carry = high & (~low);
    }
  }
}

long long int pworkARM64(signed char mid[], int mwm, signed char nonce[], int n, int *stop)
{
  uint64x2_t lmid[STATE_LENGTH], hmid[STATE_LENGTH];

  paraARM64(mid, lmid, hmid);
  int offset = HASH_LENGTH - NONCE_LENGTH;

  lmid[offset][0] = LOW00;
  lmid[offset][1] = LOW01;
  hmid[offset][0] = HIGH00;
  hmid[offset][1] = HIGH01;

  lmid[offset+1][0] = LOW10;
  lmid[offset+1][1] = LOW11;
  hmid[offset+1][0] = HIGH10;
  hmid[offset+1][1] = HIGH11;

  lmid[offset+2][0] = LOW20;
  lmid[offset+2][1] = LOW21;
  hmid[offset+2][0] = HIGH20;
  hmid[offset+2][1] = HIGH21;

  lmid[offset+3][0] = LOW30;
  lmid[offset+3][1] = LOW31;
  hmid[offset+3][0] = HIGH30;
  hmid[offset+3][1] = HIGH31;

  lmid[offset+4][0] = LOW40;
  lmid[offset+4][1] = LOW41;
  hmid[offset+4][0] = HIGH40;
  hmid[offset+4][1] = HIGH41;

  incrNARM64(n, lmid, hmid);

  return loopARM64(lmid, hmid, mwm, nonce, stop);
}
*/
import "C"
import (
	"math"
	"sync"
	"unsafe"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
)

func init() {
	proofOfWorkFuncs["CARM64"] = CARM64ProofOfWork
	proofOfWorkFuncs["SyncCARM64"] = SyncCARM64ProofOfWork
}

// CARM64ProofOfWork does proof of work on the given trytes using native C code and __int128 C type (ARM adjusted).
// This implementation follows common C standards and does not rely on SSE which is AMD64 specific.
func CARM64ProofOfWork(trytes Trytes, mwm int, parallelism ...int) (Trytes, error) {
	return cARM64ProofOfWork(trytes, mwm, nil, parallelism)
}

var syncCARM64ProofOfWork = sync.Mutex{}

// SyncCARM64ProofOfWork is like CARM64ProofOfWork() but only runs one ongoing Proof-of-Work task at a time.
func SyncCARM64ProofOfWork(trytes Trytes, mwm int, parallelism ...int) (Trytes, error) {
	syncCARM64ProofOfWork.Lock()
	defer syncCARM64ProofOfWork.Unlock()
	nonce, err := cARM64ProofOfWork(trytes, mwm, nil, parallelism...)
	if err != nil {
		return "", err
	}
	return nonce, nil
}

func cARM64ProofOfWork(trytes Trytes, mwm int, optRate chan int64, parallelism ...int) (Trytes, error) {
	if trytes == "" {
		return "", ErrInvalidTrytesForProofOfWork
	}

	tr := MustTrytesToTrits(trytes)

	c := curl.NewCurlP81().(*curl.Curl)
	c.Absorb(tr[:(TransactionTrinarySize - HashTrinarySize)])
	copy(c.State, tr[TransactionTrinarySize-HashTrinarySize:])

	numGoroutines := proofOfWorkParallelism(parallelism...)
	var result Trytes
	var rate chan int64
	if optRate != nil {
		rate = make(chan int64, numGoroutines)
	}
	exit := make(chan struct{})
	nonceChan := make(chan Trytes)

	var cancelled C.int
	for n := 0; n < numGoroutines; n++ {
		go func(n int) {
			nonce := make(Trits, NonceTrinarySize)

			r := C.pworkARM64(
				(*C.schar)(unsafe.Pointer(&c.State[0])), C.int(mwm),
				(*C.schar)(unsafe.Pointer(&nonce[0])), C.int(n), &cancelled)

			if rate != nil {
				rate <- int64(math.Abs(float64(r)))
			}

			if r >= 0 {
				select {
				case <-exit:
				case nonceChan <- MustTritsToTrytes(nonce):
					cancelled = 1
				}
			}
		}(n)
	}

	if rate != nil {
		var rateSum int64
		for i := 0; i < numGoroutines; i++ {
			rateSum += <-rate
		}
		optRate <- rateSum
	}

	result = <-nonceChan
	close(exit)
	cancelled = 1
	return result, nil
}
