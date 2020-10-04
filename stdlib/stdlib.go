/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package stdlib

import (
	"fmt"
	"strings"

	"devt.de/krotik/ecal/util"
)

/*
GetStdlibSymbols returns all available packages of stdlib and their constant
and function symbols.
*/
func GetStdlibSymbols() ([]string, []string, []string) {
	var constSymbols, funcSymbols []string
	var packageNames []string

	packageSet := make(map[string]bool)

	addSym := func(sym string, suffix string, symMap map[interface{}]interface{},
		ret []string) []string {

		if strings.HasSuffix(sym, suffix) {
			trimSym := strings.TrimSuffix(sym, suffix)
			packageSet[trimSym] = true
			for k := range symMap {
				ret = append(ret, fmt.Sprintf("%v.%v", trimSym, k))
			}
		}

		return ret
	}

	for k, v := range genStdlib {
		sym := fmt.Sprint(k)
		symMap := v.(map[interface{}]interface{})

		constSymbols = addSym(sym, "-const", symMap, constSymbols)
		funcSymbols = addSym(sym, "-func", symMap, funcSymbols)
	}
	for k := range packageSet {
		packageNames = append(packageNames, k)
	}

	return packageNames, constSymbols, funcSymbols
}

/*
GetStdlibConst looks up a constant from stdlib.
*/
func GetStdlibConst(name string) (interface{}, bool) {
	m, n := splitModuleAndName(name)

	if n != "" {
		if cmap, ok := genStdlib[fmt.Sprintf("%v-const", m)]; ok {
			if cv, ok := cmap.(map[interface{}]interface{})[n]; ok {
				return cv.(interface{}), true
			}
		}
	}

	return nil, false
}

/*
GetStdlibFunc looks up a function from stdlib.
*/
func GetStdlibFunc(name string) (util.ECALFunction, bool) {
	m, n := splitModuleAndName(name)

	if n != "" {
		if fmap, ok := genStdlib[fmt.Sprintf("%v-func", m)]; ok {
			if fn, ok := fmap.(map[interface{}]interface{})[n]; ok {
				return fn.(util.ECALFunction), true
			}
		}
	}

	return nil, false
}

/*
GetPkgDocString returns the docstring of a stdlib package.
*/
func GetPkgDocString(name string) (string, bool) {

	// TODO Implement

	return "", false
}

func splitModuleAndName(fullname string) (string, string) {
	var module, name string

	ccSplit := strings.SplitN(fullname, ".", 2)

	if len(ccSplit) != 0 {
		module = ccSplit[0]
		name = strings.Join(ccSplit[1:], "")
	}

	return module, name
}
