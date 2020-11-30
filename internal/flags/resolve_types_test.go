// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flags

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveTypes", func() {
	DescribeTable("Set() without error", func(modeStr string, exp ResolveTypes) {
		modes := ResolveTypes{}
		Expect(modes.Set(modeStr)).To(Succeed())
		Expect(modes).To(Equal(exp))
	},
		Entry("replaces", "replaces=foo", ResolveTypes{Replaces: "foo"}),
		Entry("skipRange", "skipRange='<=2.0.0 >1.0.0 || !1.1.2'", ResolveTypes{
			SkipRange: "<=2.0.0 >1.0.0 || !1.1.2",
		}),
		Entry("replaces and skipRange", "replaces=foo,skipRange='<=2.0.0 >1.0.0 || !1.1.2'", ResolveTypes{
			Replaces:  "foo",
			SkipRange: "<=2.0.0 >1.0.0 || !1.1.2",
		}),
		Entry("skipRange with double quotes", "skipRange=\"<=2.0.0 >1.0.0\"", ResolveTypes{
			SkipRange: "<=2.0.0 >1.0.0",
		}),
	)

	DescribeTable("Set() with error", func(modeStr string) {
		err := (&ResolveTypes{}).Set(modeStr)
		Expect(err).To(HaveOccurred())
	},
		Entry("empty value", ""),
		Entry("two replaces", "replaces=foo,replaces=bar"),
		Entry("two skipRanges", "skipRange=<2.0.0,skipRange=>1.0.0"),
		Entry("random resolve mode type", "foo=bar"),
		Entry("replaces with no value", "replaces="),
		Entry("skipRange with no value", "skipRange="),
		Entry("replaces and skipRange with no skipRange value", "replaces=foo,skipRange="),
		Entry("replaces and skipRange with no replaces value", "replaces=,skipRange=<1.0.0"),
	)
})
