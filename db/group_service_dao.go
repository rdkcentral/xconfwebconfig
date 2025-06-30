/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package db

// XdasCacheDao interface for XDAS cache operations
type GroupServiceCacheDao interface {
	GetGroupServiceFeatureTags(cacheKey string) map[string]string
	SetGroupServiceFeatureTags(cacheKey string, tags map[string]string) error
	DeleteGroupServiceFeatureTags(cacheKey string) error
}

type GroupServiceCacheDaoImpl struct{}

// GetGroupServiceCacheDao returns GroupServiceCacheDao
func GetGroupServiceCacheDao() GroupServiceCacheDao {
	return &GroupServiceCacheDaoImpl{}
}

// GetGroupServiceFeatureTags retrieves GroupService feature tags from the cache
func (dao GroupServiceCacheDaoImpl) GetGroupServiceFeatureTags(cacheKey string) map[string]string {
	return GetCacheManager().GetGroupServiceFeatureTags(cacheKey)
}

// SetGroupServiceFeatureTags stores GroupService feature tags in the cache
func (dao GroupServiceCacheDaoImpl) SetGroupServiceFeatureTags(cacheKey string, tags map[string]string) error {
	return GetCacheManager().SetGroupServiceFeatureTags(cacheKey, tags)
}

// DeleteGroupServiceFeatureTags removes GroupService feature tags from the cache
func (dao GroupServiceCacheDaoImpl) DeleteGroupServiceFeatureTags(cacheKey string) error {
	return GetCacheManager().DeleteGroupServiceFeatureTags(cacheKey)
}
