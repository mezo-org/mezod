package mezotime

import "time"

// Now returns the current time. It can be overridden with a mock function for
// deterministic time in unit tests.
var Now = time.Now
