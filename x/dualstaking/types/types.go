package types

import "time"

// MinLockPeriod defines the minimum lock duration for staking
const MinLockPeriod = 7 * 24 * time.Hour // 1 week

// MaxLockPeriod defines the maximum lock duration for staking
const MaxLockPeriod = 4 * 365 * 24 * time.Hour // 4 years
