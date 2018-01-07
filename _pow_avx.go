// +build avx
// +build linux,amd64

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

// #cgo LDFLAGS: -mavx
// #cgo CFLAGS: -mavx -Wall
/*
#include <stdio.h>
#include <string.h>

#ifdef _MSC_VER
#include <intrin.h>
#else
#include <x86intrin.h>
#endif

typedef union {double d; unsigned long long l;} dl;

#define HBITS ( ( (dl) 0xFFFFFFFFFFFFFFFFuLL ).d )
#define LBITS ( ( (dl) 0x0000000000000000uLL ).d )
#define HASH_LENGTH 243              //trits
#define STATE_LENGTH 3 * HASH_LENGTH //trits
#define TX_LENGTH 2673               //trytes

#define LOW00  ( ( (dl)0xDB6DB6DB6DB6DB6DuLL ).d ) //0b1101101101101101101101101101101101101101101101101101101101101101
#define HIGH00 ( ( (dl)0xB6DB6DB6DB6DB6DBuLL ).d ) //0b1011011011011011011011011011011011011011011011011011011011011011
#define LOW10  ( ( (dl)0xF1F8FC7E3F1F8FC7uLL ).d ) //0b1111000111111000111111000111111000111111000111111000111111000111
#define HIGH10 ( ( (dl)0x8FC7E3F1F8FC7E3FuLL ).d ) //0b1000111111000111111000111111000111111000111111000111111000111111
#define LOW20  ( ( (dl)0x7FFFE00FFFFC01FFuLL ).d ) //0b0111111111111111111000000000111111111111111111000000000111111111
#define HIGH20 ( ( (dl)0xFFC01FFFF803FFFFuLL ).d ) //0b1111111111000000000111111111111111111000000000111111111111111111
#define LOW30  ( ( (dl)0xFFC0000007FFFFFFuLL ).d ) //0b1111111111000000000000000000000000000111111111111111111111111111
#define HIGH30 ( ( (dl)0x003FFFFFFFFFFFFFuLL ).d ) //0b0000000000111111111111111111111111111111111111111111111111111111
#define LOW40  ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define HIGH40 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define LOW50  ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define HIGH50 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111

#define LOW01  ( ( (dl)0x6DB6DB6DB6DB6DB6uLL ).d ) //0b0110110110110110110110110110110110110110110110110110110110110110
#define HIGH01 ( ( (dl)0xDB6DB6DB6DB6DB6DuLL ).d ) //0b1101101101101101101101101101101101101101101101101101101101101101
#define LOW11  ( ( (dl)0xF8FC7E3F1F8FC7E3uLL ).d ) //0b1111100011111100011111100011111100011111100011111100011111100011
#define HIGH11 ( ( (dl)0xC7E3F1F8FC7E3F1FuLL ).d ) //0b1100011111100011111100011111100011111100011111100011111100011111
#define LOW21  ( ( (dl)0xC01FFFF803FFFF00uLL ).d ) //0b1100000000011111111111111111100000000011111111111111111100000000
#define HIGH21 ( ( (dl)0x3FFFF007FFFE00FFuLL ).d ) //0b0011111111111111111100000000011111111111111111100000000011111111
#define LOW31  ( ( (dl)0x00000FFFFFFFFFFFuLL ).d ) //0b0000000000000000000011111111111111111111111111111111111111111111
#define HIGH31 ( ( (dl)0xFFFFFFFFFFFE0000uLL ).d ) //0b1111111111111111111111111111111111111111111111100000000000000000
#define LOW41  ( ( (dl)0x000000000001FFFFuLL ).d ) //0b0000000000000000000000000000000000000000000000011111111111111111
#define HIGH41 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define LOW51  ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define HIGH51 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111

#define LOW02  ( ( (dl)0xB6DB6DB6DB6DB6DBuLL ).d ) //0b1011011011011011011011011011011011011011011011011011011011011011
#define HIGH02 ( ( (dl)0x6DB6DB6DB6DB6DB6uLL ).d ) //0b0110110110110110110110110110110110110110110110110110110110110110
#define LOW12  ( ( (dl)0xFC7E3F1F8FC7E3F1uLL ).d ) //0b1111110001111110001111110001111110001111110001111110001111110001
#define HIGH12 ( ( (dl)0xE3F1F8FC7E3F1F8FuLL ).d ) //0b1110001111110001111110001111110001111110001111110001111110001111
#define LOW22  ( ( (dl)0xFFF007FFFE00FFFFuLL ).d ) //0b1111111111110000000001111111111111111110000000001111111111111111
#define HIGH22 ( ( (dl)0xE00FFFFC01FFFF80uLL ).d ) //0b1110000000001111111111111111110000000001111111111111111110000000
#define LOW32  ( ( (dl)0x1FFFFFFFFFFFFF80uLL ).d ) //0b0001111111111111111111111111111111111111111111111111111110000000
#define HIGH32 ( ( (dl)0xFFFFFFFC0000007FuLL ).d ) //0b1111111111111111111111111111110000000000000000000000000001111111
#define LOW42  ( ( (dl)0xFFFFFFFC00000000uLL ).d ) //0b1111111111111111111111111111110000000000000000000000000000000000
#define HIGH42 ( ( (dl)0x00000003FFFFFFFFuLL ).d ) //0b0000000000000000000000000000001111111111111111111111111111111111
#define LOW52  ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define HIGH52 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111

#define LOW03  ( ( (dl)0xDB6DB6DB6DB6DB6DuLL ).d ) //0b1101101101101101101101101101101101101101101101101101101101101101
#define HIGH03 ( ( (dl)0xB6DB6DB6DB6DB6DBuLL ).d ) //0b1011011011011011011011011011011011011011011011011011011011011011
#define LOW13  ( ( (dl)0x7E3F1F8FC7E3F1F8uLL ).d ) //0b0111111000111111000111111000111111000111111000111111000111111000
#define HIGH13 ( ( (dl)0xF1F8FC7E3F1F8FC7uLL ).d ) //0b1111000111111000111111000111111000111111000111111000111111000111
#define LOW23  ( ( (dl)0x0FFFFC01FFFF803FuLL ).d ) //0b0000111111111111111111000000000111111111111111111000000000111111
#define HIGH23 ( ( (dl)0xFFF803FFFF007FFFuLL ).d ) //0b1111111111111000000000111111111111111111000000000111111111111111
#define LOW33  ( ( (dl)0xFFFFFFFFFF000000uLL ).d ) //0b1111111111111111111111111111111111111111000000000000000000000000
#define HIGH33 ( ( (dl)0xFFF8000000FFFFFFuLL ).d ) //0b1111111111111000000000000000000000000000111111111111111111111111
#define LOW43  ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111
#define HIGH43 ( ( (dl)0xFFF8000000000000uLL ).d ) //0b1111111111111000000000000000000000000000000000000000000000000000
#define LOW53  ( ( (dl)0x0007FFFFFFFFFFFFuLL ).d ) //0b0000000000000111111111111111111111111111111111111111111111111111
#define HIGH53 ( ( (dl)0xFFFFFFFFFFFFFFFFuLL ).d ) //0b1111111111111111111111111111111111111111111111111111111111111111


const int indices___[] = {
    0, 364, 728, 363, 727, 362, 726, 361, 725, 360, 724, 359, 723, 358, 722, 357, 721, 356, 720, 355, 719, 354, 718, 353, 717, 352, 716, 351, 715, 350, 714, 349, 713, 348, 712, 347, 711, 346, 710, 345, 709, 344, 708, 343, 707, 342, 706, 341, 705, 340, 704, 339, 703, 338, 702, 337, 701, 336, 700, 335, 699, 334, 698, 333, 697, 332, 696, 331, 695, 330, 694, 329, 693, 328, 692, 327, 691, 326, 690, 325, 689, 324, 688, 323, 687, 322, 686, 321, 685, 320, 684, 319, 683, 318, 682, 317, 681, 316, 680, 315, 679, 314, 678, 313, 677, 312, 676, 311, 675, 310, 674, 309, 673, 308, 672, 307, 671, 306, 670, 305, 669, 304, 668, 303, 667, 302, 666, 301, 665, 300, 664, 299, 663, 298, 662, 297, 661, 296, 660, 295, 659, 294, 658, 293, 657, 292, 656, 291, 655, 290, 654, 289, 653, 288, 652, 287, 651, 286, 650, 285, 649, 284, 648, 283, 647, 282, 646, 281, 645, 280, 644, 279, 643, 278, 642, 277, 641, 276, 640, 275, 639, 274, 638, 273, 637, 272, 636, 271, 635, 270, 634, 269, 633, 268, 632, 267, 631, 266, 630, 265, 629, 264, 628, 263, 627, 262, 626, 261, 625, 260, 624, 259, 623, 258, 622, 257, 621, 256, 620, 255, 619, 254, 618, 253, 617, 252, 616, 251, 615, 250, 614, 249, 613, 248, 612, 247, 611, 246, 610, 245, 609, 244, 608, 243, 607, 242, 606, 241, 605, 240, 604, 239, 603, 238, 602, 237, 601, 236, 600, 235, 599, 234, 598, 233, 597, 232, 596, 231, 595, 230, 594, 229, 593, 228, 592, 227, 591, 226, 590, 225, 589, 224, 588, 223, 587, 222, 586, 221, 585, 220, 584, 219, 583, 218, 582, 217, 581, 216, 580, 215, 579, 214, 578, 213, 577, 212, 576, 211, 575, 210, 574, 209, 573, 208, 572, 207, 571, 206, 570, 205, 569, 204, 568, 203, 567, 202, 566, 201, 565, 200, 564, 199, 563, 198, 562, 197, 561, 196, 560, 195, 559, 194, 558, 193, 557, 192, 556, 191, 555, 190, 554, 189, 553, 188, 552, 187, 551, 186, 550, 185, 549, 184, 548, 183, 547, 182, 546, 181, 545, 180, 544, 179, 543, 178, 542, 177, 541, 176, 540, 175, 539, 174, 538, 173, 537, 172, 536, 171, 535, 170, 534, 169, 533, 168, 532, 167, 531, 166, 530, 165, 529, 164, 528, 163, 527, 162, 526, 161, 525, 160, 524, 159, 523, 158, 522, 157, 521, 156, 520, 155, 519, 154, 518, 153, 517, 152, 516, 151, 515, 150, 514, 149, 513, 148, 512, 147, 511, 146, 510, 145, 509, 144, 508, 143, 507, 142, 506, 141, 505, 140, 504, 139, 503, 138, 502, 137, 501, 136, 500, 135, 499, 134, 498, 133, 497, 132, 496, 131, 495, 130, 494, 129, 493, 128, 492, 127, 491, 126, 490, 125, 489, 124, 488, 123, 487, 122, 486, 121, 485, 120, 484, 119, 483, 118, 482, 117, 481, 116, 480, 115, 479, 114, 478, 113, 477, 112, 476, 111, 475, 110, 474, 109, 473, 108, 472, 107, 471, 106, 470, 105, 469, 104, 468, 103, 467, 102, 466, 101, 465, 100, 464, 99, 463, 98, 462, 97, 461, 96, 460, 95, 459, 94, 458, 93, 457, 92, 456, 91, 455, 90, 454, 89, 453, 88, 452, 87, 451, 86, 450, 85, 449, 84, 448, 83, 447, 82, 446, 81, 445, 80, 444, 79, 443, 78, 442, 77, 441, 76, 440, 75, 439, 74, 438, 73, 437, 72, 436, 71, 435, 70, 434, 69, 433, 68, 432, 67, 431, 66, 430, 65, 429, 64, 428, 63, 427, 62, 426, 61, 425, 60, 424, 59, 423, 58, 422, 57, 421, 56, 420, 55, 419, 54, 418, 53, 417, 52, 416, 51, 415, 50, 414, 49, 413, 48, 412, 47, 411, 46, 410, 45, 409, 44, 408, 43, 407, 42, 406, 41, 405, 40, 404, 39, 403, 38, 402, 37, 401, 36, 400, 35, 399, 34, 398, 33, 397, 32, 396, 31, 395, 30, 394, 29, 393, 28, 392, 27, 391, 26, 390, 25, 389, 24, 388, 23, 387, 22, 386, 21, 385, 20, 384, 19, 383, 18, 382, 17, 381, 16, 380, 15, 379, 14, 378, 13, 377, 12, 376, 11, 375, 10, 374, 9, 373, 8, 372, 7, 371, 6, 370, 5, 369, 4, 368, 3, 367, 2, 366, 1, 365, 0};

void transform256(__m256d *lmid, __m256d *hmid)
{
  __m256d one = _mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
  int j, r, t1, t2;
  __m256d alpha, beta, gamma, delta,ngamma;
  __m256d *lto = lmid + STATE_LENGTH, *hto = hmid + STATE_LENGTH;
  __m256d *lfrom = lmid, *hfrom = hmid;
  for (r = 0; r < 26; r++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      t1 = indices___[j];
      t2 = indices___[j + 1];

      alpha = lfrom[t1];
      beta = hfrom[t1];
      gamma = hfrom[t2];
      ngamma = _mm256_andnot_pd(gamma,one);
     delta =  _mm256_and_pd(_mm256_or_pd(alpha,ngamma), _mm256_xor_pd(lfrom[t2],beta)); //(alpha | (~gamma)) & (lfrom[t2] ^ beta);


      lto[j] = _mm256_andnot_pd(delta,one);  //~delta;
      hto[j] = _mm256_or_pd(_mm256_xor_pd(alpha,gamma),delta); //(alpha ^ gamma) | delta;
    }
    __m256d *lswap = lfrom, *hswap = hfrom;
    lfrom = lto;
    hfrom = hto;
    lto = lswap;
    hto = hswap;
  }
  for (j = 0; j < HASH_LENGTH; j++)
  {
    t1 = indices___[j];
    t2 = indices___[j + 1];

    alpha = lfrom[t1];
    beta = hfrom[t1];
    gamma = hfrom[t2];
      ngamma = _mm256_andnot_pd(gamma,one);
     delta =  _mm256_and_pd(_mm256_or_pd(alpha,ngamma), _mm256_xor_pd(lfrom[t2],beta)); //(alpha | (~gamma)) & (lfrom[t2] ^ beta);


      lto[j] = _mm256_andnot_pd(delta,one);  //~delta;
      hto[j] = _mm256_or_pd(_mm256_xor_pd(alpha,gamma),delta); //(alpha ^ gamma) | delta;
  }
}

int incr256(__m256d *mid_low, __m256d *mid_high)
{
  int i;
  __m256d carry;
  carry = _mm256_set_pd(LBITS, LBITS,LBITS,LBITS);
  for (i = 6; i < HASH_LENGTH && (i == 6 || carry[0]); i++)
  {
    __m256d low = mid_low[i], high = mid_high[i];
    mid_low[i] = _mm256_xor_pd(high, low);
    mid_high[i] = low;
    carry =  _mm256_andnot_pd(low,high); //high & (~low);
  }
  return i == HASH_LENGTH;
}

void seri256(__m256d *low, __m256d *high, int n, char *r)
{
  int i = 0, index = 0;
  if (n > 63 && n<128)
  {
    n -= 64;
    index = 1;
  }
  if (n >= 128 && n<192)
  {
    n -= 128;
    index = 2;
  }
  if (n >= 192 && n<256)
  {
    n -=  192;
    index = 3;
  }
  for (i = 0; i < HASH_LENGTH; i++)
  {
    long long l= ((dl)low[i][index]).l;
    long long h= ((dl)high[i][index]).l;
    long ll = (l >> n) & 1;
    long hh = (h >> n) & 1;
    if (hh == 0 && ll == 1)
    {
      r[i] = -1;
    }
    if (hh == 1 && ll == 1)
    {
      r[i] = 0;
    }
    if (hh == 1 && ll == 0)
    {
      r[i] = 1;
    }
  }
}

int check256(__m256d *l, __m256d *h, int m)
{
  int i, j; //omit init for speed

  __m256d nonce_probe = _mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
  for (i = HASH_LENGTH - m; i < HASH_LENGTH; i++)
  {
    nonce_probe =_mm256_andnot_pd(_mm256_xor_pd(l[i],h[i]),nonce_probe);   //&= ~(l[i] ^ h[i]);
    if (nonce_probe[0] == LBITS && nonce_probe[1] == LBITS && nonce_probe[2] == LBITS && nonce_probe[3] == LBITS)
    {
      return -1;
    }
  }
  for (j = 0; j < 3; j++)
  {
    for (i = 0; i < 64; i++)
    {
      long long np= ((dl)nonce_probe[j]).l;
      if ( (np >> i) & 1)
      {
        return i + j * 64;
      }
    }
  }
  return -2;
}

int loop256(__m256d *lmid, __m256d *hmid, int m, char *nonce,int *stop)
{
  int i = 0, n = 0, j = 0;

  __m256d lcpy[STATE_LENGTH * 2], hcpy[STATE_LENGTH * 2];
  for (i = 0; !incr256(lmid, hmid) && !*stop; i++)
  {
    for (j = 0; j < STATE_LENGTH; j++)
    {
      lcpy[j] = lmid[j];
      hcpy[j] = hmid[j];
    }
    transform256(lcpy, hcpy);
    if ((n = check256(lcpy + STATE_LENGTH, hcpy + STATE_LENGTH, m)) >= 0)
    {
      seri256(lmid, hmid, n, nonce);
      return i * 256;
    }
  }
  return -i*256-1;
}

// 01:-1 11:0 10:1
void para256(char in[], __m256d l[], __m256d h[])
{
  int i = 0;
  for (i = 0; i < STATE_LENGTH; i++)
  {
    switch (in[i])
    {
    case 0:
      l[i] = _mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
      h[i] = _mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
      break;
    case 1:
      l[i] = _mm256_set_pd(LBITS, LBITS,LBITS,LBITS);
      h[i] = _mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
      break;
    case -1:
      l[i] =_mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
      h[i] =  _mm256_set_pd(LBITS, LBITS,LBITS,LBITS);
      break;
    }
  }
}

void incrN256(int n,__m256d *mid_low, __m256d *mid_high)
{
  int i,j;
  for (j=0;j<n;j++){
    __m256d carry;
    carry =_mm256_set_pd(HBITS, HBITS,HBITS,HBITS);
    for (i = HASH_LENGTH * 2 / 3; i < HASH_LENGTH &&  carry[0]; i++)
    {
      __m256d low = mid_low[i], high = mid_high[i];
      mid_low[i] = _mm256_xor_pd(high, low);
      mid_high[i] = low;
      carry = _mm256_andnot_pd(low,high);// high & (~low);
    }
  }
}

int pwork256(char mid[], int mwm, char nonce[],int n,int *stop)
{
  __m256d lmid[STATE_LENGTH], hmid[STATE_LENGTH];

  para256(mid, lmid, hmid);
  lmid[0] = _mm256_set_pd(LOW00, LOW01,LOW02,LOW03);
  hmid[0] = _mm256_set_pd(HIGH00, HIGH01,HIGH02,HIGH03);
  lmid[1] = _mm256_set_pd(LOW10, LOW11,LOW12,LOW13);
  hmid[1] = _mm256_set_pd(HIGH10,HIGH11,HIGH12,HIGH13);
  lmid[2] = _mm256_set_pd(LOW20, LOW21,LOW22,LOW23);
  hmid[2] = _mm256_set_pd(HIGH20,HIGH21,HIGH32,HIGH23);
  lmid[3] = _mm256_set_pd(LOW30, LOW31,LOW32,LOW33);
  hmid[3] = _mm256_set_pd(HIGH30,HIGH31,HIGH32,HIGH33);
  lmid[4] = _mm256_set_pd(LOW40, LOW41,LOW42,LOW43);
  hmid[4] = _mm256_set_pd(HIGH40,HIGH41,HIGH42,HIGH43);
  lmid[5] = _mm256_set_pd(LOW50, LOW51,LOW52,LOW53);
  hmid[5] = _mm256_set_pd(HIGH50,HIGH51,HIGH52,HIGH53);

	incrN256(n, lmid, hmid);
  return loop256(lmid, hmid, mwm, nonce,stop);
}
*/
import "C"
import (
	"sync"
	"unsafe"
)

func init() {
	pows["PowAVX"] = PowAVX
}

var countAVX int64

// PowAVX is proof of work of iota for amd64 using AVX.
func PowAVX(trytes Trytes, mwm int) (Trytes, error) {
	countAVX = 0
	c := NewCurl()
	c.Absorb(trytes[:(transactionTrinarySize-HashSize)/3])

	var (
		stop   int64
		result Trytes
		wg     sync.WaitGroup
		mutex  sync.Mutex
	)

	for n := 0; n < PowProcs; n++ {
		wg.Add(1)
		go func(n int) {
			nonce := make(Trits, HashSize)

			// nolint: gas
			r := C.pwork256((*C.char)(
				unsafe.Pointer(&c.state[0])), C.int(mwm), (*C.char)(unsafe.Pointer(&nonce[0])),
				C.int(n), (*C.int)(unsafe.Pointer(&stop)))

			mutex.Lock()

			switch {
			case r >= 0:
				result = nonce.Trytes()
				stop = 1
				countAVX += int64(r)
			default:
				countAVX += int64(-r + 1)
			}

			mutex.Unlock()
			wg.Done()
		}(n)
	}
	wg.Wait()
	return result, nil
}
