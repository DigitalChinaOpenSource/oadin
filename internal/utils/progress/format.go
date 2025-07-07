//*****************************************************************************
// Copyright 2025 Intel Corporation
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
//*****************************************************************************

package progress

import (
	"fmt"
	"math"
	"strconv"

	"intel.com/aog/internal/constants"
)

func HumanNumber(b uint64) string {
	switch {
	case b >= constants.Billion:
		number := float64(b) / constants.Billion
		if number == math.Floor(number) {
			return fmt.Sprintf("%.0fB", number) // no decimals if whole number
		}
		return fmt.Sprintf("%.1fB", number) // one decimal if not a whole number
	case b >= constants.Million:
		number := float64(b) / constants.Million
		if number == math.Floor(number) {
			return fmt.Sprintf("%.0fM", number) // no decimals if whole number
		}
		return fmt.Sprintf("%.2fM", number) // two decimals if not a whole number
	case b >= constants.Thousand:
		return fmt.Sprintf("%.0fK", float64(b)/constants.Thousand)
	default:
		return strconv.FormatUint(b, 10)
	}
}

func HumanBytes(b int64) string {
	var value float64
	var unit string

	switch {
	case b >= constants.TeraByte:
		value = float64(b) / constants.TeraByte
		unit = "TB"
	case b >= constants.GigaByte:
		value = float64(b) / constants.GigaByte
		unit = "GB"
	case b >= constants.MegaByte:
		value = float64(b) / constants.MegaByte
		unit = "MB"
	case b >= constants.KiloByte:
		value = float64(b) / constants.KiloByte
		unit = "KB"
	default:
		return fmt.Sprintf("%d B", b)
	}

	switch {
	case value >= 100:
		return fmt.Sprintf("%d %s", int(value), unit)
	case value >= 10:
		return fmt.Sprintf("%d %s", int(value), unit)
	case value != math.Trunc(value):
		return fmt.Sprintf("%.1f %s", value, unit)
	default:
		return fmt.Sprintf("%d %s", int(value), unit)
	}
}

func HumanBytes2(b uint64) string {
	switch {
	case b >= constants.GibiByte:
		return fmt.Sprintf("%.1f GiB", float64(b)/constants.GibiByte)
	case b >= constants.MebiByte:
		return fmt.Sprintf("%.1f MiB", float64(b)/constants.MebiByte)
	case b >= constants.KibiByte:
		return fmt.Sprintf("%.1f KiB", float64(b)/constants.KibiByte)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
