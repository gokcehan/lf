// Copyright 2015 The Tops'l Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package views

// Alignment represents the alignment of an object, and consists of
// either or both of horizontal and vertical alignment.
type Alignment int

const (
	// HAlignLeft indicates alignment on the left edge.
	HAlignLeft Alignment = 1 << iota

	// HAlignCenter indicates horizontally centered.
	HAlignCenter

	// HAlignRight indicates alignment on the right edge.
	HAlignRight

	// VAlignTop indicates alignment on the top edge.
	VAlignTop

	// VAlignCenter indicates vertically centered.
	VAlignCenter

	// VAlignBottom indicates alignment on the bottom edge.
	VAlignBottom
)
const (
	// AlignBegin indicates alignment at the top left corner.
	AlignBegin = HAlignLeft | VAlignTop

	// AlignEnd indicates alignment at the bottom right corner.
	AlignEnd = HAlignRight | VAlignBottom

	// AlignMiddle indicates full centering.
	AlignMiddle = HAlignCenter | VAlignCenter
)

// Orientation represents the direction of a widget or layout.
type Orientation int

const (
	// Horizontal indicates left to right orientation.
	Horizontal = iota

	// Vertical indicates top to bottom orientation.
	Vertical
)
