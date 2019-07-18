/*
 * Copyright (c) 2018 IOTA Stiftung
 * https://github.com/iotaledger/entangled
 *
 * Refer to the LICENSE file for licensing information
 */

#ifndef __UTILS_CONTAINERS_trit_t_mam_msg_read_context_t_MAP_H__
#define __UTILS_CONTAINERS_trit_t_mam_msg_read_context_t_MAP_H__

#include <stdbool.h>

#include "uthash.h"
#include "common/errors.h"
#include "common/trinary/trits.h"
#include "mam/mam/message.h"

/*
 * This Generic map allows mapping any key type to any value type
 * assuming that key can be trivially copied, to allow for
 * user-defined types, user should add dependency in "map_generator.bzl"
 * and include the required files in this header file
 */

#ifdef __cplusplus
extern "C" {
#endif

typedef struct trit_t_to_mam_msg_read_context_t_map_entry_s {
  trit_t *key;
  mam_msg_read_context_t *value;
  UT_hash_handle hh;
} trit_t_to_mam_msg_read_context_t_map_entry_t;

typedef struct trit_t_to_mam_msg_read_context_t_map_s {
  size_t key_size;
  size_t value_size;
  trit_t_to_mam_msg_read_context_t_map_entry_t* map;
} trit_t_to_mam_msg_read_context_t_map_t;

retcode_t trit_t_to_mam_msg_read_context_t_map_init(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                              size_t const key_size, size_t const value_size);

size_t trit_t_to_mam_msg_read_context_t_map_size(trit_t_to_mam_msg_read_context_t_map_t const *const map);

retcode_t trit_t_to_mam_msg_read_context_t_map_add(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                             trit_t const *const key,
                                             mam_msg_read_context_t const *const value);

bool trit_t_to_mam_msg_read_context_t_map_contains(trit_t_to_mam_msg_read_context_t_map_t const *const map,
                                             trit_t const *const key);

bool trit_t_to_mam_msg_read_context_t_map_find(trit_t_to_mam_msg_read_context_t_map_t const *const map,
                                         trit_t const *const key,
                                         trit_t_to_mam_msg_read_context_t_map_entry_t **const res);

retcode_t trit_t_to_mam_msg_read_context_t_map_free(trit_t_to_mam_msg_read_context_t_map_t *const map);

bool trit_t_to_mam_msg_read_context_t_map_cmp(trit_t_to_mam_msg_read_context_t_map_t const *const lhs,
                                        trit_t_to_mam_msg_read_context_t_map_t const *const rhs);

bool trit_t_to_mam_msg_read_context_t_map_remove(trit_t_to_mam_msg_read_context_t_map_t *const map,
trit_t const *const key);

retcode_t trit_t_to_mam_msg_read_context_t_map_remove_entry(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                      trit_t_to_mam_msg_read_context_t_map_entry_t *const entry);

#ifdef __cplusplus
}
#endif

#endif  // __UTILS_CONTAINERS_trit_t_mam_msg_read_context_t_MAP_H__
