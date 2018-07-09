// +build linux,darwin,windows amd64 linux,arm64

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

// #cgo LDFLAGS:
// #cgo CFLAGS: -Wall
/*
#include <stdio.h>
#include <string.h>
#include <stdint.h>

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

const int indices_C128[] = {
    0, 364, 728, 363, 727, 362, 726, 361, 725, 360, 724, 359, 723, 358, 722, 357, 721, 356, 720, 355, 719, 354, 718, 353, 717, 352, 716, 351, 715, 350, 714, 349, 713, 348, 712, 347, 711, 346, 710, 345, 709, 344, 708, 343, 707, 342, 706, 341, 705, 340, 704, 339, 703, 338, 702, 337, 701, 336, 700, 335, 699, 334, 698, 333, 697, 332, 696, 331, 695, 330, 694, 329, 693, 328, 692, 327, 691, 326, 690, 325, 689, 324, 688, 323, 687, 322, 686, 321, 685, 320, 684, 319, 683, 318, 682, 317, 681, 316, 680, 315, 679, 314, 678, 313, 677, 312, 676, 311, 675, 310, 674, 309, 673, 308, 672, 307, 671, 306, 670, 305, 669, 304, 668, 303, 667, 302, 666, 301, 665, 300, 664, 299, 663, 298, 662, 297, 661, 296, 660, 295, 659, 294, 658, 293, 657, 292, 656, 291, 655, 290, 654, 289, 653, 288, 652, 287, 651, 286, 650, 285, 649, 284, 648, 283, 647, 282, 646, 281, 645, 280, 644, 279, 643, 278, 642, 277, 641, 276, 640, 275, 639, 274, 638, 273, 637, 272, 636, 271, 635, 270, 634, 269, 633, 268, 632, 267, 631, 266, 630, 265, 629, 264, 628, 263, 627, 262, 626, 261, 625, 260, 624, 259, 623, 258, 622, 257, 621, 256, 620, 255, 619, 254, 618, 253, 617, 252, 616, 251, 615, 250, 614, 249, 613, 248, 612, 247, 611, 246, 610, 245, 609, 244, 608, 243, 607, 242, 606, 241, 605, 240, 604, 239, 603, 238, 602, 237, 601, 236, 600, 235, 599, 234, 598, 233, 597, 232, 596, 231, 595, 230, 594, 229, 593, 228, 592, 227, 591, 226, 590, 225, 589, 224, 588, 223, 587, 222, 586, 221, 585, 220, 584, 219, 583, 218, 582, 217, 581, 216, 580, 215, 579, 214, 578, 213, 577, 212, 576, 211, 575, 210, 574, 209, 573, 208, 572, 207, 571, 206, 570, 205, 569, 204, 568, 203, 567, 202, 566, 201, 565, 200, 564, 199, 563, 198, 562, 197, 561, 196, 560, 195, 559, 194, 558, 193, 557, 192, 556, 191, 555, 190, 554, 189, 553, 188, 552, 187, 551, 186, 550, 185, 549, 184, 548, 183, 547, 182, 546, 181, 545, 180, 544, 179, 543, 178, 542, 177, 541, 176, 540, 175, 539, 174, 538, 173, 537, 172, 536, 171, 535, 170, 534, 169, 533, 168, 532, 167, 531, 166, 530, 165, 529, 164, 528, 163, 527, 162, 526, 161, 525, 160, 524, 159, 523, 158, 522, 157, 521, 156, 520, 155, 519, 154, 518, 153, 517, 152, 516, 151, 515, 150, 514, 149, 513, 148, 512, 147, 511, 146, 510, 145, 509, 144, 508, 143, 507, 142, 506, 141, 505, 140, 504, 139, 503, 138, 502, 137, 501, 136, 500, 135, 499, 134, 498, 133, 497, 132, 496, 131, 495, 130, 494, 129, 493, 128, 492, 127, 491, 126, 490, 125, 489, 124, 488, 123, 487, 122, 486, 121, 485, 120, 484, 119, 483, 118, 482, 117, 481, 116, 480, 115, 479, 114, 478, 113, 477, 112, 476, 111, 475, 110, 474, 109, 473, 108, 472, 107, 471, 106, 470, 105, 469, 104, 468, 103, 467, 102, 466, 101, 465, 100, 464, 99, 463, 98, 462, 97, 461, 96, 460, 95, 459, 94, 458, 93, 457, 92, 456, 91, 455, 90, 454, 89, 453, 88, 452, 87, 451, 86, 450, 85, 449, 84, 448, 83, 447, 82, 446, 81, 445, 80, 444, 79, 443, 78, 442, 77, 441, 76, 440, 75, 439, 74, 438, 73, 437, 72, 436, 71, 435, 70, 434, 69, 433, 68, 432, 67, 431, 66, 430, 65, 429, 64, 428, 63, 427, 62, 426, 61, 425, 60, 424, 59, 423, 58, 422, 57, 421, 56, 420, 55, 419, 54, 418, 53, 417, 52, 416, 51, 415, 50, 414, 49, 413, 48, 412, 47, 411, 46, 410, 45, 409, 44, 408, 43, 407, 42, 406, 41, 405, 40, 404, 39, 403, 38, 402, 37, 401, 36, 400, 35, 399, 34, 398, 33, 397, 32, 396, 31, 395, 30, 394, 29, 393, 28, 392, 27, 391, 26, 390, 25, 389, 24, 388, 23, 387, 22, 386, 21, 385, 20, 384, 19, 383, 18, 382, 17, 381, 16, 380, 15, 379, 14, 378, 13, 377, 12, 376, 11, 375, 10, 374, 9, 373, 8, 372, 7, 371, 6, 370, 5, 369, 4, 368, 3, 367, 2, 366, 1, 365, 0};

unsigned long long low64(unsigned __int128 *var) 
{
  return (unsigned long long) *var; 
}

unsigned long long high64(unsigned __int128 *var) 
{
  return (unsigned long long) (*var >> 64); 
}

unsigned long long get64(unsigned __int128 *var, int index) 
{

  if (index == 0) {
    return (unsigned long long) *var;
    } 
  return (unsigned long long) (*var >> 64); 
}

unsigned __int128 set_int128_2x64(unsigned long long a, unsigned long long b)
{
  unsigned __int128 r;

  r = (unsigned __int128) b;
  r |= (unsigned __int128) a << 64;

  return r;
}

void transformC128(unsigned __int128 *lmid, unsigned __int128 *hmid)
{
  int j, r, t1, t2;
  unsigned __int128 alpha, beta, gamma, delta;
  unsigned __int128 *lto = lmid + STATE_LENGTH, *hto = hmid + STATE_LENGTH;
  unsigned __int128 *lfrom = lmid, *hfrom = hmid;

  for (r = 0; r < 80; r++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      t1 = indices_C128[j];
      t2 = indices_C128[j + 1];
      alpha = lfrom[t1];
      beta = hfrom[t1];
      gamma = hfrom[t2];
      delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);
      lto[j] = ~delta;
      hto[j] = (alpha ^ gamma) | delta;
    }
    unsigned __int128 *lswap = lfrom, *hswap = hfrom;
    lfrom = lto;
    hfrom = hto;
    lto = lswap;
    hto = hswap;
  }

  for (j = 0; j < HASH_LENGTH; j++)
  {
    t1 = indices_C128[j];
    t2 = indices_C128[j + 1];
    alpha = lfrom[t1];
    beta = hfrom[t1];
    gamma = hfrom[t2];
    delta = (alpha | (~gamma)) & (lfrom[t2] ^ beta);
    lto[j] = ~delta;
    hto[j] = (alpha ^ gamma) | delta;
  }
}

int incrC128(unsigned __int128 *mid_low, unsigned __int128 *mid_high)
{
  int i;
  unsigned __int128 carry;

  carry = set_int128_2x64(LOW00, LOW01);

  for (i = INCR_START; i < HASH_LENGTH && (i == INCR_START || low64(&carry)); i++)
  {
    unsigned __int128 low = mid_low[i], high = mid_high[i];
    mid_low[i] = high ^ low;
    mid_high[i] = low;
    carry = high & (~low);
  }
  return i == HASH_LENGTH;
}

void seriC128(unsigned __int128 *low, unsigned __int128 *high, int n, signed char *r)
{
  int i = 0, index = 0;

  if (n > 63)
  {
    n -= 64;
    index = 1;
  }

  for (i = HASH_LENGTH-NONCE_LENGTH; i < HASH_LENGTH; i++)
  {
    unsigned long long ll = (get64(&low[i], index) >> n) & 1;
    unsigned long long hh = (get64(&high[i], index) >> n) & 1;

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

int checkC128(unsigned __int128 *l, unsigned __int128 *h, int m)
{
  int i, j; //omit init for speed
  unsigned __int128 nonce_probe = set_int128_2x64(HBITS, HBITS);

  for (i = HASH_LENGTH - m; i < HASH_LENGTH; i++)
  {
    nonce_probe &= ~(l[i] ^ h[i]);

    if (low64(&nonce_probe) == LBITS && high64(&nonce_probe) == LBITS)
    {
      return -1;
    }
  }

  for (j = 0; j < 2; j++)
  {
    for (i = 0; i < 64; i++)
    {
      if ((get64(&nonce_probe, j) >> i) & 1)
      {
        return i + j * 64;
      }
    }
  }

  return -2;
}

int stopC128 = 1;

long long int loopC128(unsigned __int128 *lmid, unsigned __int128 *hmid, int m, signed char *nonce)
{
  int n = 0, j = 0;
  long long int i = 0;
  unsigned __int128 lcpy[STATE_LENGTH * 2], hcpy[STATE_LENGTH * 2];

  for (i = 0; !incrC128(lmid, hmid) && !stopC128; i++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      lcpy[j] = lmid[j];
      hcpy[j] = hmid[j];
    }
    transformC128(lcpy, hcpy);

    if ((n = checkC128(lcpy + STATE_LENGTH, hcpy + STATE_LENGTH, m)) >= 0)
    {
      seriC128(lmid, hmid, n, nonce);
      return i * 128;
    }
  }
  return -i*128-1;
}

void paraC128(signed char in[], unsigned __int128 l[], unsigned __int128 h[])
{
  int i = 0;
  for (i = 0; i < STATE_LENGTH; i++)
  {
    switch (in[i])
    {
    case 0:
      l[i] = set_int128_2x64(HBITS, HBITS);
      h[i] = set_int128_2x64(HBITS, HBITS);
      break;
    case 1:
      l[i] = set_int128_2x64(LBITS, LBITS);
      h[i] = set_int128_2x64(HBITS, HBITS);
      break;
    case -1:
      l[i] = set_int128_2x64(HBITS, HBITS);
      h[i] = set_int128_2x64(LBITS, LBITS);      
      break;
    }
  }
}

void incrNC128(int n,unsigned __int128 *mid_low, unsigned __int128 *mid_high)
{
  int i,j;
  for (j=0;j<n;j++){
    unsigned __int128 carry;
    carry = set_int128_2x64(HBITS, HBITS);
    for (i = HASH_LENGTH * 2/3 + 4; i < HASH_LENGTH * 2/3 + 4 + 27 && low64(&carry); i++)
    {
      unsigned __int128 low = mid_low[i], high = mid_high[i];
      mid_low[i] = high ^ low;
      mid_high[i] = low;
      carry = high & (~low);
    }
  }
}

long long int pworkC128(signed char mid[], int mwm, signed char nonce[], int n)
{
  unsigned __int128 lmid[STATE_LENGTH], hmid[STATE_LENGTH];
  paraC128(mid, lmid, hmid);
  int offset = HASH_LENGTH - NONCE_LENGTH;
 
  lmid[offset] = set_int128_2x64(LOW00, LOW01);
  hmid[offset] = set_int128_2x64(HIGH00, HIGH01);
  lmid[offset+1] = set_int128_2x64(LOW10, LOW11);
  hmid[offset+1] = set_int128_2x64(HIGH10, HIGH11);
  lmid[offset+2] = set_int128_2x64(LOW20, LOW21);
  hmid[offset+2] = set_int128_2x64(HIGH20, HIGH21);
  lmid[offset+3] = set_int128_2x64(LOW30, LOW31);
  hmid[offset+3] = set_int128_2x64(HIGH30, HIGH31);
  lmid[offset+4] = set_int128_2x64(LOW40, LOW41);
  hmid[offset+4] = set_int128_2x64(HIGH40, HIGH41);
  
  incrNC128(n, lmid, hmid);
  return loopC128(lmid, hmid, mwm, nonce);
}
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

func init() {
	powFuncs["PowC128"] = PowC128
}

var countC128 int64

// PowC128 is a proof of work library for Iota that uses the standard __int128 C type that is available in 64 bit processors (AMD64 and ARM64).
// This PoW calculator follows common C standards and does not rely on SSE which is AMD64 specific.
func PowC128(trytes Trytes, mwm int) (Trytes, error) {
	if C.stopC128 == 0 {
		C.stopC128 = 1
		return "", errors.New("pow is already running, stopped")
	}

	if trytes == "" {
		return "", errors.New("invalid trytes")
	}

	C.stopC128 = 0
	countC128 = 0
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
			r := C.pworkC128((*C.schar)(unsafe.Pointer(&c.state[0])), C.int(mwm), (*C.schar)(unsafe.Pointer(&nonce[0])), C.int(n))
			mutex.Lock()

			switch {
			case r >= 0:
				result = nonce.Trytes()
				C.stopC128 = 1
				countC128 += int64(r)
			default:
				countC128 += int64(-r + 1)
			}

			mutex.Unlock()
			wg.Done()
		}(n)
	}

	wg.Wait()
	C.stopC128 = 1
	return result, nil
}
