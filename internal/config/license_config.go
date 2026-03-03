package config

import "time"

var LicenseGracePeriodDays int = 7

func LicenseGracePeriod() time.Duration {
	return time.Duration(LicenseGracePeriodDays) * 24 * time.Hour
}
