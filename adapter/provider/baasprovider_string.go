// Code generated by "stringer -type=BaaSProvider"; DO NOT EDIT

package provider

import "fmt"

const _BaaSProvider_name = "HyperledgerFabricKaleidoAmazonManagedBlockchainHyperledgerCelloIBMMicrosoftSAPShieldCureBlockoHyperledgerIrohaHyperledgerIndyHyperledgerSawtooth"

var _BaaSProvider_index = [...]uint8{0, 17, 24, 47, 63, 66, 75, 78, 88, 94, 110, 125, 144}

func (i BaaSProvider) String() string {
	if i < 0 || i >= BaaSProvider(len(_BaaSProvider_index)-1) {
		return fmt.Sprintf("BaaSProvider(%d)", i)
	}
	return _BaaSProvider_name[_BaaSProvider_index[i]:_BaaSProvider_index[i+1]]
}
