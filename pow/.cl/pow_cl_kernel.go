// +build cgo
// +build gpu

package cl

var kernel = `
#define HASH_LENGTH 243
#define STATE_LENGTH 3 * HASH_LENGTH
#define HALF_LENGTH 364
#define HIGH_BITS 0xFFFFFFFFFFFFFFFF
#define LOW_BITS 0x0000000000000000
//#define HIGH_BITS 0b1111111111111111111111111111111111111111111111111111111111111111L
//#define LOW_BITS 0b0000000000000000000000000000000000000000000000000000000000000000L

typedef long trit_t;

void increment(__global trit_t *mid_low,
		__global trit_t *mid_high,
		__private size_t from_index,
		__private size_t to_index);
void copy_mid_to_state(
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__global trit_t *state_low,
		__global trit_t *state_high,
		__private size_t id,
		__private size_t l_size,
		__private size_t l_trits);
void transform(__global trit_t *state_low,
		__global trit_t *state_high,
		__private size_t id,
		__private size_t l_size,
		__private size_t l_trits);
void check(__global trit_t *state_low,
		__global trit_t *state_high,
		__global volatile char *found,
		__constant size_t *min_weight_magnitude,
		__global trit_t *nonce_probe,
		__private size_t gr_id);
void setup_ids(
		__private size_t *id,
		__private size_t *gid,
		__private size_t *gr_id,
		__private size_t *l_size,
		__private size_t *n_trits
		);

void increment(
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__private size_t from_index,
		__private size_t to_index
		) {
	size_t i;
	for (i = from_index; i < to_index; i++) {
		if (mid_low[i] == (trit_t)LOW_BITS) {
			mid_low[i] = (trit_t)HIGH_BITS;
			mid_high[i] = (trit_t)LOW_BITS;
		} else {
			if (mid_high[i] == (trit_t)LOW_BITS) {
				mid_high[i] = (trit_t)HIGH_BITS;
			} else {
				mid_low[i] = (trit_t)LOW_BITS;
			}
			break;
		}
	}
}

void copy_mid_to_state(
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__global trit_t *state_low,
		__global trit_t *state_high,
		__private size_t id,
		__private size_t l_size,
		__private size_t n_trits
		) {
	size_t i, j;
	for(i = 0; i < n_trits; i++) {
		j = id + i*l_size;
		state_low[j] = mid_low[j];
		state_high[j] = mid_high[j];
	}
}

void transform(
		__global trit_t *state_low,
		__global trit_t *state_high,
		__private size_t id,
		__private size_t l_size,
		__private size_t n_trits
		) {
	__private size_t round, i, j, t1, t2;
	__private trit_t alpha, beta, gamma, delta, sp_low[3], sp_high[3];
	for(round = 0; round < 27; round++) {
		for(i = 0; i < n_trits; i++) {
			j = id + i*l_size;
			t1 = j == 0? 0:(((j - 1)%2)+1)*HALF_LENGTH - ((j-1)>>1);
			t2 = ((j%2)+1)*HALF_LENGTH - ((j)>>1);

			alpha = state_low[t1];
			beta = state_high[t1];
			gamma = state_high[t2];
			delta = (alpha | (~gamma)) & (state_low[t2] ^ beta);

			sp_low[i] = ~delta;
			sp_high[i] = (alpha ^ gamma) | delta;
		}
		barrier(CLK_LOCAL_MEM_FENCE);
		for(i = 0; i < n_trits; i++) {
			j = id + i*l_size;
			state_low[j] = sp_low[i];
			state_high[j] = sp_high[i];
		}
		barrier(CLK_LOCAL_MEM_FENCE);
	}
}

void check(
		__global trit_t *state_low,
		__global trit_t *state_high,
		__global volatile char *found,
		__constant size_t *min_weight_magnitude,
		__global trit_t *nonce_probe,
		__private size_t gr_id
		) {
	int i;
	*nonce_probe = HIGH_BITS;
	for (i = HASH_LENGTH - *min_weight_magnitude; i < HASH_LENGTH; i++) {
		*nonce_probe &= ~(state_low[i] ^ state_high[i]);
		if(*nonce_probe == 0) return;
	}
	if(*nonce_probe != 0) {
		//*nonce_probe = 1 << __builtin_ctzl(*nonce_probe);
		*found = gr_id + 1;
	}

}

void setup_ids(
		__private size_t *id,
		__private size_t *gid,
		__private size_t *gr_id,
		__private size_t *l_size,
		__private size_t *n_trits
		) {
	__private size_t l_rem;
	*id = get_local_id(0);
	*l_size = get_local_size(0);
	*gr_id = get_global_id(0)/ *l_size;
	*gid = *gr_id*STATE_LENGTH;
	l_rem = STATE_LENGTH % *l_size; 
	*n_trits = STATE_LENGTH/ *l_size;
	*n_trits += l_rem == 0? 0: 1;
	*n_trits -= (*n_trits) * (*id) < STATE_LENGTH ? 0 : 1;
}

__kernel void init (
		__global trit_t *trit_hash,
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__global trit_t *state_low,
		__global trit_t *state_high,
		__constant size_t *min_weight_magnitude,
		__global volatile char *found,
		__global trit_t *nonce_probe,
		__constant size_t *loop_count
		) {
	__private size_t i, j, id, gid, gr_id, gl_off, l_size, n_trits;
	setup_ids(&id, &gid, &gr_id, &l_size, &n_trits);
	gl_off = get_global_offset(0);
	
	if(id == 0 && gr_id == 0) {
		*found = 0;
	}

	if(gr_id == 0) return;

	for(i = 0; i < n_trits; i++) {
		j = id + i*l_size;
		mid_low[gid + j] = mid_low[j];
		mid_high[gid + j] = mid_high[j];
	}

	if(id == 0) {
		for(i = 0; i < gr_id + gl_off; i++) {
			increment(&(mid_low[gid]), &(mid_high[gid]), HASH_LENGTH / 3, (HASH_LENGTH / 3) * 2);
		}
	}
}

__kernel void search (
		__global trit_t *trit_hash,
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__global trit_t *state_low,
		__global trit_t *state_high,
		__constant size_t *min_weight_magnitude,
		__global volatile char *found,
		__global trit_t *nonce_probe,
		__constant size_t *loop_count
		) {
	__private size_t i, id, gid, gr_id, l_size, n_trits;
	setup_ids(&id, &gid, &gr_id, &l_size, &n_trits);

	for(i = 0; i < *loop_count; i++) {
		if(id == 0) increment(&(mid_low[gid]), &(mid_high[gid]), (HASH_LENGTH/3)*2, HASH_LENGTH);

		barrier(CLK_LOCAL_MEM_FENCE);
		copy_mid_to_state(&(mid_low[gid]), &(mid_high[gid]), &(state_low[gid]), &(state_high[gid]), id, l_size, n_trits);

		barrier(CLK_LOCAL_MEM_FENCE);
		transform(&(state_low[gid]), &(state_high[gid]), id, l_size, n_trits);

		barrier(CLK_LOCAL_MEM_FENCE);
		if(id == 0) check(&(state_low[gid]), &(state_high[gid]), found, min_weight_magnitude, &(nonce_probe[gr_id]), gr_id);

		barrier(CLK_LOCAL_MEM_FENCE);
		if(*found != 0) break;
	}
}


__kernel void finalize (
		__global trit_t *trit_hash,
		__global trit_t *mid_low,
		__global trit_t *mid_high,
		__global trit_t *state_low,
		__global trit_t *state_high,
		__constant size_t *min_weight_magnitude,
		__global volatile char *found,
		__global trit_t *nonce_probe,
		__constant size_t *loop_count
		) {
	__private size_t i,j, id, gid, gr_id, l_size, n_trits;
	setup_ids(&id, &gid, &gr_id, &l_size, &n_trits);

	if(gr_id == (size_t)(*found - 1) && nonce_probe[gr_id] != 0) {
		for(i = 0; i < n_trits; i++) {
			j = id + i*l_size;
			if(j < HASH_LENGTH) {
				trit_hash[j] = (mid_low[gid + j] & nonce_probe[gr_id]) == 0 ? 
					1 : (mid_high[gid + j] & nonce_probe[gr_id]) == 0 ? -1 : 0;
			}
		}
	}
}
`
