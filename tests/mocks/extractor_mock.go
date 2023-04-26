/*
Copyright © 2022 - 2023 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mocks

import v1 "github.com/rancher/elemental-cli/pkg/types/v1"

type FakeImageExtractor struct {
	Logger     v1.Logger
	SideEffect func(imageRef, destination, platformRef string, local bool) error
}

var _ v1.ImageExtractor = FakeImageExtractor{}

func NewFakeImageExtractor(logger v1.Logger) *FakeImageExtractor {
	return &FakeImageExtractor{
		Logger: logger,
	}
}

func (f FakeImageExtractor) ExtractImage(imageRef, destination, platformRef string, local bool) error {
	f.Logger.Debugf("extracting %s to %s in platform %s", imageRef, destination, platformRef)
	if f.SideEffect != nil {
		f.Logger.Debugf("running sideeffect")
		return f.SideEffect(imageRef, destination, platformRef, local)
	}

	return nil
}
