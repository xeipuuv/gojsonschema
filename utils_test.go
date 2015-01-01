// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           janmentzel
// author-github    https://github.com/janmentzel
// author-mail      ? ( forward to xeipuuv@gmail.com )
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description     (Unit) Tests for utils ( Float / Integer conversion ).
//
// created          08-08-2013

package gojsonschema

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestIsFloat64IntegerA(t *testing.T) {

	assert.False(t, isFloat64AnInteger(math.Inf(+1)))
	assert.False(t, isFloat64AnInteger(math.Inf(-1)))
	assert.False(t, isFloat64AnInteger(math.NaN()))
	assert.False(t, isFloat64AnInteger(math.Float64frombits((1<<11-1)<<52|1)))
	assert.True(t, isFloat64AnInteger(float64(0.0)))
	assert.True(t, isFloat64AnInteger(-float64(0.0)))
	assert.False(t, isFloat64AnInteger(float64(0.5)))
	assert.True(t, isFloat64AnInteger(float64(1.0)))
	assert.True(t, isFloat64AnInteger(-float64(1.0)))
	assert.False(t, isFloat64AnInteger(float64(1.1)))
	assert.True(t, isFloat64AnInteger(float64(131.0)))
	assert.True(t, isFloat64AnInteger(-float64(131.0)))
	assert.True(t, isFloat64AnInteger(float64(1<<52-1)))
	assert.True(t, isFloat64AnInteger(-float64(1<<52-1)))
	assert.True(t, isFloat64AnInteger(float64(1<<52)))
	assert.True(t, isFloat64AnInteger(-float64(1<<52)))
	assert.True(t, isFloat64AnInteger(float64(1<<53-1)))
	assert.True(t, isFloat64AnInteger(-float64(1<<53-1)))
	assert.True(t, isFloat64AnInteger(float64(1<<53-1)))
	assert.False(t, isFloat64AnInteger(float64(1<<53)))
	assert.False(t, isFloat64AnInteger(-float64(1<<53)))
	assert.False(t, isFloat64AnInteger(float64(1<<63)))
	assert.False(t, isFloat64AnInteger(-float64(1<<63)))
	assert.False(t, isFloat64AnInteger(math.Nextafter(float64(1<<63), math.MaxFloat64)))
	assert.False(t, isFloat64AnInteger(-math.Nextafter(float64(1<<63), math.MaxFloat64)))
	assert.False(t, isFloat64AnInteger(float64(1<<70+3<<21)))
	assert.False(t, isFloat64AnInteger(-float64(1<<70+3<<21)))

	assert.False(t, isFloat64AnInteger(math.Nextafter(float64(9007199254740991.0), math.MaxFloat64)))
	assert.True(t, isFloat64AnInteger(float64(9007199254740991.0)))

	assert.True(t, isFloat64AnInteger(float64(-9007199254740991.0)))
	assert.False(t, isFloat64AnInteger(math.Nextafter(float64(-9007199254740991.0), -math.MaxFloat64)))
}

func TestIsFloat64Integer(t *testing.T) {
	// fails. MaxUint64 is to large for JS anyway, so we can ignore it here.
	//assert.True(t, isFloat64AnInteger(float64(math.MaxUint64)))

	assert.False(t, isFloat64AnInteger(math.MaxInt64))
	assert.False(t, isFloat64AnInteger(1<<62))
	assert.False(t, isFloat64AnInteger(math.MinInt64))
	assert.True(t, isFloat64AnInteger(100100100100))
	assert.True(t, isFloat64AnInteger(-100100100100))
	assert.True(t, isFloat64AnInteger(100100100))
	assert.True(t, isFloat64AnInteger(-100100100))
	assert.True(t, isFloat64AnInteger(100100))
	assert.True(t, isFloat64AnInteger(-100100))
	assert.True(t, isFloat64AnInteger(100))
	assert.True(t, isFloat64AnInteger(-100))
	assert.True(t, isFloat64AnInteger(-0))
	assert.True(t, isFloat64AnInteger(-1))
	assert.True(t, isFloat64AnInteger(1))
	assert.True(t, isFloat64AnInteger(0))
	assert.True(t, isFloat64AnInteger(77))
	assert.True(t, isFloat64AnInteger(-77))
	assert.True(t, isFloat64AnInteger(1e10))
	assert.True(t, isFloat64AnInteger(-1e10))
	assert.True(t, isFloat64AnInteger(100100100.0))
	assert.True(t, isFloat64AnInteger(-100100100.0))

	assert.False(t, isFloat64AnInteger(100100100100.1))
	assert.False(t, isFloat64AnInteger(-100100100100.1))
	assert.False(t, isFloat64AnInteger(math.MaxFloat64))
	assert.False(t, isFloat64AnInteger(-math.MaxFloat64))
	assert.False(t, isFloat64AnInteger(1.1))
	assert.False(t, isFloat64AnInteger(-1.1))
	assert.False(t, isFloat64AnInteger(1.000000000001))
	assert.False(t, isFloat64AnInteger(-1.000000000001))
	assert.False(t, isFloat64AnInteger(1e-10))
	assert.False(t, isFloat64AnInteger(-1e-10))

	assert.False(t, isFloat64AnInteger(1.001))
	assert.False(t, isFloat64AnInteger(-1.001))

	assert.False(t, isFloat64AnInteger(0.0001))

	assert.False(t, isFloat64AnInteger(1<<62))
	assert.False(t, isFloat64AnInteger(math.MinInt64))
	assert.False(t, isFloat64AnInteger(math.MaxInt64))
	assert.False(t, isFloat64AnInteger(-1<<62))

	assert.False(t, isFloat64AnInteger(1e-10))
	assert.False(t, isFloat64AnInteger(-1e-10))

	assert.False(t, isFloat64AnInteger(1.000000000001))
	assert.False(t, isFloat64AnInteger(-1.000000000001))

	assert.False(t, isFloat64AnInteger(4.611686018427387904e7))
	assert.False(t, isFloat64AnInteger(-4.611686018427387904e7))
}

func TestResultErrorFormatNumber(t *testing.T) {
	assert.Equal(t, "1", resultErrorFormatNumber(1))
	assert.Equal(t, "-1", resultErrorFormatNumber(-1))
	assert.Equal(t, "0", resultErrorFormatNumber(0))
	// unfortunately, can not be recognized as float
	assert.Equal(t, "0", resultErrorFormatNumber(0.0))

	assert.Equal(t, "1.001", resultErrorFormatNumber(1.001))
	assert.Equal(t, "-1.001", resultErrorFormatNumber(-1.001))
	assert.Equal(t, "0.0001", resultErrorFormatNumber(0.0001))

	// casting math.MaxInt64 (1<<63 -1) to float back to int64
	// becomes negative. obviousely because of bit missinterpretation.
	// so simply test a slightly smaller "large" integer here
	assert.Equal(t, "4.611686018427388e+18", resultErrorFormatNumber(1<<62))
	// with negative int64 max works
	assert.Equal(t, "-9.223372036854776e+18", resultErrorFormatNumber(math.MinInt64))
	assert.Equal(t, "-4.611686018427388e+18", resultErrorFormatNumber(-1<<62))

	assert.Equal(t, "10000000000", resultErrorFormatNumber(1e10))
	assert.Equal(t, "-10000000000", resultErrorFormatNumber(-1e10))

	assert.Equal(t, "1.000000000001", resultErrorFormatNumber(1.000000000001))
	assert.Equal(t, "-1.000000000001", resultErrorFormatNumber(-1.000000000001))
	assert.Equal(t, "1e-10", resultErrorFormatNumber(1e-10))
	assert.Equal(t, "-1e-10", resultErrorFormatNumber(-1e-10))
	assert.Equal(t, "4.6116860184273876e+07", resultErrorFormatNumber(4.611686018427387904e7))
	assert.Equal(t, "-4.6116860184273876e+07", resultErrorFormatNumber(-4.611686018427387904e7))

}
