/*
 * Copyright (c) 2018 IOTA Stiftung
 * https://github.com/iotaledger/entangled
 *
 * Refer to the LICENSE file for licensing information
 */

#include "mam/api/trit_t_to_mam_msg_read_context_t_map.h"

retcode_t trit_t_to_mam_msg_read_context_t_map_init(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                              size_t const key_size, size_t const value_size) {
  map->key_size = key_size;
  map->value_size = value_size;
  map->map = NULL;

  return RC_OK;
}

size_t trit_t_to_mam_msg_read_context_t_map_size(trit_t_to_mam_msg_read_context_t_map_t const *const map) {
  return HASH_COUNT(map->map);
}

retcode_t trit_t_to_mam_msg_read_context_t_map_add(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                             trit_t const *const key,
                                             mam_msg_read_context_t const *const value) {
  trit_t_to_mam_msg_read_context_t_map_entry_t *map_entry = NULL;
  map_entry = (trit_t_to_mam_msg_read_context_t_map_entry_t *)malloc(
      sizeof(trit_t_to_mam_msg_read_context_t_map_entry_t));

  if (map_entry == NULL) {
    return RC_OOM;
  }

  if ((map_entry->key = (trit_t*)malloc(map->key_size)) == NULL) {
    return RC_OOM;
  }

  if ((map_entry->value = (mam_msg_read_context_t*)malloc(map->value_size)) == NULL) {
    return RC_OOM;
  }

  memcpy(map_entry->key, key, map->key_size);
  memcpy(map_entry->value, value, map->value_size);
  HASH_ADD_KEYPTR(hh, map->map, map_entry->key, map->key_size, map_entry);

  return RC_OK;
}

bool trit_t_to_mam_msg_read_context_t_map_contains(trit_t_to_mam_msg_read_context_t_map_t const *const map,
                                             trit_t const *const key) {
  trit_t_to_mam_msg_read_context_t_map_entry_t *entry = NULL;

  if (map == NULL || map->map == NULL) {
    return false;
  }

  HASH_FIND(hh, map->map, key,map->key_size, entry);

  return entry != NULL;
}

bool trit_t_to_mam_msg_read_context_t_map_find(trit_t_to_mam_msg_read_context_t_map_t const *const map,
                                         trit_t const *const key,
                                         trit_t_to_mam_msg_read_context_t_map_entry_t **const res) {
  if (map == NULL || map->map == NULL) {
    return false;
  }
  if (res == NULL) {
    return RC_NULL_PARAM;
  }

  HASH_FIND(hh, map->map, key, map->key_size, *res);

  return *res != NULL;
}

retcode_t trit_t_to_mam_msg_read_context_t_map_free(trit_t_to_mam_msg_read_context_t_map_t *const map) {
  trit_t_to_mam_msg_read_context_t_map_entry_t *curr_entry = NULL;
  trit_t_to_mam_msg_read_context_t_map_entry_t *tmp_entry = NULL;

  HASH_ITER(hh, map->map, curr_entry, tmp_entry) {
    free(curr_entry->key);
    free(curr_entry->value);
    HASH_DEL(map->map, curr_entry);
    free(curr_entry);
  }

  map->map = NULL;
  return RC_OK;
}

bool trit_t_to_mam_msg_read_context_t_map_cmp(trit_t_to_mam_msg_read_context_t_map_t const *const lhs,
trit_t_to_mam_msg_read_context_t_map_t const *const rhs){

  if (HASH_COUNT(lhs->map) != HASH_COUNT(rhs->map)){
    return false;
  }

  trit_t_to_mam_msg_read_context_t_map_entry_t *curr_entry = NULL;
  trit_t_to_mam_msg_read_context_t_map_entry_t *tmp_entry = NULL;

  HASH_ITER(hh, lhs->map, curr_entry, tmp_entry) {
  if (!trit_t_to_mam_msg_read_context_t_map_contains(rhs,curr_entry->key)){
    return false;
    }
  }
  return true;
}

bool trit_t_to_mam_msg_read_context_t_map_remove(trit_t_to_mam_msg_read_context_t_map_t *const map,
trit_t const *const key) {
trit_t_to_mam_msg_read_context_t_map_entry_t *entry = NULL;

  if (map == NULL || map->map == NULL) {
    return false;
  }

  HASH_FIND(hh, map->map, key,map->key_size, entry);

  trit_t_to_mam_msg_read_context_t_map_remove_entry(map, entry);

  return entry != NULL;
}

retcode_t trit_t_to_mam_msg_read_context_t_map_remove_entry(trit_t_to_mam_msg_read_context_t_map_t *const map,
                                      trit_t_to_mam_msg_read_context_t_map_entry_t *const entry) {
  if (entry != NULL) {
    free(entry->key);
    free(entry->value);
    HASH_DEL(map->map, entry);
    free(entry);
  }

  return RC_OK;
}
