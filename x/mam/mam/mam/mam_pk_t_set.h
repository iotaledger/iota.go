/*
 * Copyright (c) 2018 IOTA Stiftung
 * https://github.com/iotaledger/entangled
 *
 * Refer to the LICENSE file for licensing information
 */

#ifndef __UTILS_CONTAINERS_SET_mam_pk_t_SET_H__
#define __UTILS_CONTAINERS_SET_mam_pk_t_SET_H__

#include "uthash.h"

#include <inttypes.h>
#include <stdbool.h>
#include "common/errors.h"
#include "mam/mam/channel.h"

#define SET_ITER(set, entry, tmp) HASH_ITER(hh, set, entry, tmp)

#ifdef __cplusplus
extern "C" {
#endif

typedef struct mam_pk_t_set_entry_s {
  mam_pk_t value;
  UT_hash_handle hh;
} mam_pk_t_set_entry_t;

typedef mam_pk_t_set_entry_t *mam_pk_t_set_t;

typedef retcode_t (*mam_pk_t_on_container_func)(void *container,
                                                mam_pk_t *type);

size_t mam_pk_t_set_size(mam_pk_t_set_t const set);
retcode_t mam_pk_t_set_add(mam_pk_t_set_t *const set,
                            mam_pk_t const *const value);
retcode_t mam_pk_t_set_remove(mam_pk_t_set_t *const set,
                                mam_pk_t const *const value);
retcode_t mam_pk_t_set_remove_entry(mam_pk_t_set_t *const set,
                                      mam_pk_t_set_entry_t *const entry);
retcode_t mam_pk_t_set_append(mam_pk_t_set_t const *const set1,
                               mam_pk_t_set_t *const set2);
bool mam_pk_t_set_contains(mam_pk_t_set_t const *const set,
                            mam_pk_t const *const value);
bool mam_pk_t_set_find(mam_pk_t_set_t const *const set,
        mam_pk_t const *const , mam_pk_t_set_entry_t const ** entry);
void mam_pk_t_set_free(mam_pk_t_set_t *const set);
retcode_t mam_pk_t_set_for_each(mam_pk_t_set_t const *const set,
                                 mam_pk_t_on_container_func func,
                                 void *const container);

bool mam_pk_t_set_cmp(mam_pk_t_set_t const * const lhs, mam_pk_t_set_t const * const rhs);

#ifdef __cplusplus
}
#endif

#endif  // __UTILS_CONTAINERS_SET_mam_pk_t_SET_H__
