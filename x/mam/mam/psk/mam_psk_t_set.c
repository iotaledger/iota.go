/*
 * Copyright (c) 2018 IOTA Stiftung
 * https://github.com/iotaledger/entangled
 *
 * Refer to the LICENSE file for licensing information
 */

#include "mam/psk/mam_psk_t_set.h"

size_t mam_psk_t_set_size(mam_psk_t_set_t const set) {
  return HASH_COUNT(set);
}

retcode_t mam_psk_t_set_add(mam_psk_t_set_t *const set,
                            mam_psk_t const *const value) {
  mam_psk_t_set_entry_t *entry = NULL;

  if (!mam_psk_t_set_contains(set, value)) {
    if ((entry = (mam_psk_t_set_entry_t *)malloc(sizeof(mam_psk_t_set_entry_t))) == NULL) {
      return RC_OOM;
    }
    memcpy(&entry->value, value, sizeof(mam_psk_t));
    HASH_ADD(hh, *set, value, sizeof(mam_psk_t), entry);
  }
  return RC_OK;
}

retcode_t mam_psk_t_set_remove(mam_psk_t_set_t *const set,
                                mam_psk_t const *const value) {
  mam_psk_t_set_entry_t *entry = NULL;

  if (set != NULL && * set != NULL &&  value != NULL) {
    HASH_FIND(hh, *set, value, sizeof(mam_psk_t), entry);
    return mam_psk_t_set_remove_entry(set, entry);
  }
  return RC_OK;
}

retcode_t mam_psk_t_set_remove_entry(mam_psk_t_set_t *const set,
                                      mam_psk_t_set_entry_t * const entry) {
  if (set != NULL && * set != NULL && entry != NULL) {
    HASH_DEL(*set, entry);
    free(entry);
  }
  return RC_OK;
}

retcode_t mam_psk_t_set_append(mam_psk_t_set_t const *const set1,
                               mam_psk_t_set_t *const set2) {
  retcode_t ret = RC_OK;
  mam_psk_t_set_entry_t *iter = NULL, *tmp = NULL;

  HASH_ITER(hh, *set1, iter, tmp) {
    if ((ret = mam_psk_t_set_add(set2, &iter->value)) != RC_OK) {
      return ret;
    }
  }
  return ret;
}

bool mam_psk_t_set_contains(mam_psk_t_set_t const *const set,
                            mam_psk_t const *const value) {
  mam_psk_t_set_entry_t *entry = NULL;

  if (*set == NULL) {
    return false;
  }

  HASH_FIND(hh, *set, value, sizeof(mam_psk_t), entry);
  return entry != NULL;
}



bool mam_psk_t_set_find(mam_psk_t_set_t const *const set,
        mam_psk_t const *const value, mam_psk_t_set_entry_t const ** entry){
if (*set == NULL) {
return false;
}

if (entry == NULL){
  return RC_NULL_PARAM;
}

HASH_FIND(hh, *set, value, sizeof(mam_psk_t), *entry);
return *entry != NULL;

}

void mam_psk_t_set_free(mam_psk_t_set_t *const set) {
  mam_psk_t_set_entry_t *iter = NULL, *tmp = NULL;

  if (set == NULL || *set == NULL) {
    return;
  }

  HASH_ITER(hh, *set, iter, tmp) {
    HASH_DEL(*set, iter);
    free(iter);
  }
  *set = NULL;
}

retcode_t mam_psk_t_set_for_each(mam_psk_t_set_t const *const set,
                                 mam_psk_t_on_container_func func,
                                 void *const container) {
  retcode_t ret = RC_OK;
  mam_psk_t_set_entry_t *curr_entry = NULL;
  mam_psk_t_set_entry_t *tmp_entry = NULL;

  HASH_ITER(hh, *set, curr_entry, tmp_entry) {
    if ((ret = func(container, &curr_entry->value)) != RC_OK) {
      return ret;
    }
  }
  return ret;
}

bool mam_psk_t_set_cmp(mam_psk_t_set_t const *const lhs,
                          mam_psk_t_set_t const *const rhs){

  if (HASH_COUNT(*lhs) != HASH_COUNT(*rhs)){
    return false;
  }

  mam_psk_t_set_entry_t *curr_entry = NULL;
  mam_psk_t_set_entry_t *tmp_entry = NULL;

  HASH_ITER(hh, *lhs, curr_entry, tmp_entry) {
      if (!(mam_psk_t_set_contains(rhs, &curr_entry->value))){
        return false;
      }
  }
  return true;
}
