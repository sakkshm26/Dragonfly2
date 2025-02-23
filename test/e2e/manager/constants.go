/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package manager

const (
	cdnCachePath = "/tmp/cdn/download"

	managerService = "dragonfly-manager.dragonfly-system.svc"
	managerPort    = "8080"
	preheatPath    = "api/v1/jobs"
	managerTag     = "d7y/manager"

	dragonflyNamespace = "dragonfly-system"
	e2eNamespace       = "dragonfly-e2e"

	proxy            = "localhost:65001"
	hostnameFilePath = "/etc/hostname"
)
