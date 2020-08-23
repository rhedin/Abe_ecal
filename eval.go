/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

/*
Package eval contains the main API for the event condition language ECAL.
*/
package eval

import "devt.de/krotik/ecal/util"

// TODO: Maybe API documentation - access comments during runtime

/*
processor is the main implementation for the Processor interface.
*/
type processor struct {
	// TODO: GM GraphManager is part of initial values published in the global scope

	util.Logger
}
