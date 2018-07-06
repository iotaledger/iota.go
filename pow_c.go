// +build cgo
// +build linux darwin windows

/*
MIT License

Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

// #cgo CFLAGS: -Wall
/*
#include <stdio.h>
#include <string.h>

#define HBITS 0xFFFFFFFFFFFFFFFFL
#define LBITS 0x0000000000000000L
#define HASH_LENGTH 243              //trits
#define NONCE_LENGTH 81              //trits
#define STATE_LENGTH 3 * HASH_LENGTH //trits
#define TX_LENGTH 2673               //trytes
#define INCR_START HASH_LENGTH - NONCE_LENGTH + 4 + 27

#define LOW0 0xDB6DB6DB6DB6DB6DL
#define HIGH0 0xB6DB6DB6DB6DB6DBL
#define LOW1 0xF1F8FC7E3F1F8FC7L
#define HIGH1 0x8FC7E3F1F8FC7E3FL
#define LOW2 0x7FFFE00FFFFC01FFL
#define HIGH2 0xFFC01FFFF803FFFFL
#define LOW3 0xFFC0000007FFFFFFL
#define HIGH3 0x003FFFFFFFFFFFFFL

const int indices_[] = {
    0, 364, 728, 363, 727, 362, 726, 361, 725, 360, 724, 359, 723, 358, 722, 357, 721, 356, 720, 355, 719, 354, 718, 353, 717, 352, 716, 351, 715, 350, 714, 349, 713, 348, 712, 347, 711, 346, 710, 345, 709, 344, 708, 343, 707, 342, 706, 341, 705, 340, 704, 339, 703, 338, 702, 337, 701, 336, 700, 335, 699, 334, 698, 333, 697, 332, 696, 331, 695, 330, 694, 329, 693, 328, 692, 327, 691, 326, 690, 325, 689, 324, 688, 323, 687, 322, 686, 321, 685, 320, 684, 319, 683, 318, 682, 317, 681, 316, 680, 315, 679, 314, 678, 313, 677, 312, 676, 311, 675, 310, 674, 309, 673, 308, 672, 307, 671, 306, 670, 305, 669, 304, 668, 303, 667, 302, 666, 301, 665, 300, 664, 299, 663, 298, 662, 297, 661, 296, 660, 295, 659, 294, 658, 293, 657, 292, 656, 291, 655, 290, 654, 289, 653, 288, 652, 287, 651, 286, 650, 285, 649, 284, 648, 283, 647, 282, 646, 281, 645, 280, 644, 279, 643, 278, 642, 277, 641, 276, 640, 275, 639, 274, 638, 273, 637, 272, 636, 271, 635, 270, 634, 269, 633, 268, 632, 267, 631, 266, 630, 265, 629, 264, 628, 263, 627, 262, 626, 261, 625, 260, 624, 259, 623, 258, 622, 257, 621, 256, 620, 255, 619, 254, 618, 253, 617, 252, 616, 251, 615, 250, 614, 249, 613, 248, 612, 247, 611, 246, 610, 245, 609, 244, 608, 243, 607, 242, 606, 241, 605, 240, 604, 239, 603, 238, 602, 237, 601, 236, 600, 235, 599, 234, 598, 233, 597, 232, 596, 231, 595, 230, 594, 229, 593, 228, 592, 227, 591, 226, 590, 225, 589, 224, 588, 223, 587, 222, 586, 221, 585, 220, 584, 219, 583, 218, 582, 217, 581, 216, 580, 215, 579, 214, 578, 213, 577, 212, 576, 211, 575, 210, 574, 209, 573, 208, 572, 207, 571, 206, 570, 205, 569, 204, 568, 203, 567, 202, 566, 201, 565, 200, 564, 199, 563, 198, 562, 197, 561, 196, 560, 195, 559, 194, 558, 193, 557, 192, 556, 191, 555, 190, 554, 189, 553, 188, 552, 187, 551, 186, 550, 185, 549, 184, 548, 183, 547, 182, 546, 181, 545, 180, 544, 179, 543, 178, 542, 177, 541, 176, 540, 175, 539, 174, 538, 173, 537, 172, 536, 171, 535, 170, 534, 169, 533, 168, 532, 167, 531, 166, 530, 165, 529, 164, 528, 163, 527, 162, 526, 161, 525, 160, 524, 159, 523, 158, 522, 157, 521, 156, 520, 155, 519, 154, 518, 153, 517, 152, 516, 151, 515, 150, 514, 149, 513, 148, 512, 147, 511, 146, 510, 145, 509, 144, 508, 143, 507, 142, 506, 141, 505, 140, 504, 139, 503, 138, 502, 137, 501, 136, 500, 135, 499, 134, 498, 133, 497, 132, 496, 131, 495, 130, 494, 129, 493, 128, 492, 127, 491, 126, 490, 125, 489, 124, 488, 123, 487, 122, 486, 121, 485, 120, 484, 119, 483, 118, 482, 117, 481, 116, 480, 115, 479, 114, 478, 113, 477, 112, 476, 111, 475, 110, 474, 109, 473, 108, 472, 107, 471, 106, 470, 105, 469, 104, 468, 103, 467, 102, 466, 101, 465, 100, 464, 99, 463, 98, 462, 97, 461, 96, 460, 95, 459, 94, 458, 93, 457, 92, 456, 91, 455, 90, 454, 89, 453, 88, 452, 87, 451, 86, 450, 85, 449, 84, 448, 83, 447, 82, 446, 81, 445, 80, 444, 79, 443, 78, 442, 77, 441, 76, 440, 75, 439, 74, 438, 73, 437, 72, 436, 71, 435, 70, 434, 69, 433, 68, 432, 67, 431, 66, 430, 65, 429, 64, 428, 63, 427, 62, 426, 61, 425, 60, 424, 59, 423, 58, 422, 57, 421, 56, 420, 55, 419, 54, 418, 53, 417, 52, 416, 51, 415, 50, 414, 49, 413, 48, 412, 47, 411, 46, 410, 45, 409, 44, 408, 43, 407, 42, 406, 41, 405, 40, 404, 39, 403, 38, 402, 37, 401, 36, 400, 35, 399, 34, 398, 33, 397, 32, 396, 31, 395, 30, 394, 29, 393, 28, 392, 27, 391, 26, 390, 25, 389, 24, 388, 23, 387, 22, 386, 21, 385, 20, 384, 19, 383, 18, 382, 17, 381, 16, 380, 15, 379, 14, 378, 13, 377, 12, 376, 11, 375, 10, 374, 9, 373, 8, 372, 7, 371, 6, 370, 5, 369, 4, 368, 3, 367, 2, 366, 1, 365, 0};

void transform64(unsigned long *lmid, unsigned long *hmid)
{
  int j, r, t1, t2;
  unsigned long alpha, beta, gamma, delta;
  unsigned long *lfrom = lmid, *hfrom = hmid;
  unsigned long *lto = lmid + STATE_LENGTH, *hto = hmid + STATE_LENGTH;

  for (r = 0; r < 80; r++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      t1 = indices_[j];
      t2 = indices_[j + 1];

      alpha = lfrom[t1];
      beta = hfrom[t1];
      gamma = hfrom[t2];
      delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);

      lto[j] = ~delta;
      hto[j] = (alpha ^ gamma) | delta;
    }
    unsigned long *lswap = lfrom, *hswap = hfrom;
    lfrom = lto;
    hfrom = hto;
    lto = lswap;
    hto = hswap;
  }

  for (j = 0; j < HASH_LENGTH; j++)
  {
    t1 = indices_[j];
    t2 = indices_[j + 1];

    alpha = lfrom[t1];
    beta = hfrom[t1];
    gamma = hfrom[t2];
    delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);

    lto[j] = ~delta; //6
    hto[j] = (alpha ^ gamma) | delta;
  }
}

int incr(unsigned long *mid_low, unsigned long *mid_high)
{
  int i;
  unsigned long carry = 1;
  for (i = INCR_START; i < HASH_LENGTH && carry; i++)
  {
    unsigned long low = mid_low[i], high = mid_high[i];
    mid_low[i] = high ^ low;
    mid_high[i] = low;
    carry = high & (~low);
  }
  return i == HASH_LENGTH;
}

void seri(unsigned long *l, unsigned long *h, int n, signed char *r)
{
  int i = 0;
  for (i = HASH_LENGTH - NONCE_LENGTH; i < HASH_LENGTH; i++)
  {
    int ll = (l[i] >> n) & 1;
    int hh = (h[i] >> n) & 1;
    if (hh == 0 && ll == 1)
    {
      r[i-HASH_LENGTH + NONCE_LENGTH] = -1;
    }
    if (hh == 1 && ll == 1)
    {
      r[i-HASH_LENGTH + NONCE_LENGTH] = 0;
    }
    if (hh == 1 && ll == 0)
    {
      r[i-HASH_LENGTH + NONCE_LENGTH] = 1;
    }
  }
}

int check(unsigned long *l, unsigned long *h, int m)
{
  int i;
  unsigned long nonce_probe = HBITS;
  for (i = HASH_LENGTH - m; i < HASH_LENGTH; i++)
  {
    nonce_probe &= ~(l[i] ^ h[i]);
    if (nonce_probe == 0)
      return -1;
  }
  for (i = 0; i < 64; i++)
  {
    if ((nonce_probe >> i) & 1)
    {
      return i;
    }
  }
  return -1;
}

int stopC=1;

long long int loop_cpu(unsigned long *lmid, unsigned long *hmid, int m, signed char *nonce)
{
  int n = 0;
  long long int i = 0;
  unsigned long lcpy[STATE_LENGTH * 2], hcpy[STATE_LENGTH * 2];

  for (i = 0; !incr(lmid, hmid) && !stopC; i++)
  {
    memcpy(lcpy, lmid, STATE_LENGTH * sizeof(long));
    memcpy(hcpy, hmid, STATE_LENGTH * sizeof(long));
    transform64(lcpy, hcpy);
    if ((n = check(lcpy + STATE_LENGTH, hcpy + STATE_LENGTH, m)) >= 0)
    {
      seri(lmid, hmid, n, nonce);
      return i * 64;
    }
  }
  return -i*64+1;
}

// 01:-1 11:0 10:1
void para(signed char in[], unsigned long l[], unsigned long h[])
{
  int i = 0;
  for (i = 0; i < STATE_LENGTH; i++)
  {
    switch (in[i])
    {
    case 0:
      l[i] = HBITS;
      h[i] = HBITS;
      break;
    case 1:
      l[i] = LBITS;
      h[i] = HBITS;
      break;
    case -1:
      l[i] = HBITS;
      h[i] = LBITS;
      break;
    }
  }
}

void incrN(int n,unsigned long *mid_low, unsigned long *mid_high)
{
  int i,j;
  for (j=0;j<n;j++){
    unsigned long carry = 1;
    for (i = INCR_START - 27; i < INCR_START && carry; i++)
    {
      unsigned long low = mid_low[i], high = mid_high[i];
      mid_low[i] = high ^ low;
      mid_high[i] = low;
      carry = high & (~low);
    }
  }
}


long long int pwork(signed char mid[], int mwm, signed char nonce[],int n)
{
  unsigned long lmid[STATE_LENGTH] = {0}, hmid[STATE_LENGTH] = {0};

  para(mid, lmid, hmid);
  int offset = HASH_LENGTH - NONCE_LENGTH;
  lmid[offset] = LOW0;
  hmid[offset] = HIGH0;
  lmid[offset+1] = LOW1;
  hmid[offset+1] = HIGH1;
  lmid[offset+2] = LOW2;
  hmid[offset+2] = HIGH2;
  lmid[offset+3] = LOW3;
  hmid[offset+3] = HIGH3;

	incrN(n, lmid, hmid);
  return loop_cpu(lmid, hmid, mwm, nonce);
}
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

func init() {
	powFuncs["PowC"] = PowC
}

var countC int64

// PowC is proof of work of iota using pure C.
func PowC(trytes Trytes, mwm int) (Trytes, error) {
	if C.stopC == 0 {
		C.stopC = 1
		return "", errors.New("pow is already running, stopped")
	}

	if trytes == "" {
		return "", errors.New("invalid trytes")
	}
	C.stopC = 0
	countC = 0

	c := NewCurl()
	c.Absorb(trytes[:(TransactionTrinarySize-HashSize)/3])
	tr := trytes.Trits()
	copy(c.state, tr[TransactionTrinarySize-HashSize:])

	var (
		result Trytes
		wg     sync.WaitGroup
		mutex  sync.Mutex
	)

	for n := 0; n < PowProcs; n++ {
		wg.Add(1)
		go func(n int) {
			nonce := make(Trits, NonceTrinarySize)

			// nolint: gas
			r := C.pwork((*C.schar)(unsafe.Pointer(
				&c.state[0])), C.int(mwm), (*C.schar)(unsafe.Pointer(&nonce[0])), C.int(n))

			mutex.Lock()

			switch {
			case r >= 0:
				result = nonce.Trytes()
				C.stopC = 1
				countC += int64(r)
			default:
				countC += int64(-r + 1)
			}

			mutex.Unlock()
			wg.Done()
		}(n)
	}

	wg.Wait()
	C.stopC = 1
	return result, nil
}
