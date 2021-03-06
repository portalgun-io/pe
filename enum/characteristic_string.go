// Code generated by "stringer -trimprefix Characteristic -type Characteristic"; DO NOT EDIT.

package enum

import "strconv"

const _Characteristic_name = "RelocsStrippedExecutableImageLineNumsStrippedLocalSymsStrippedAggressiveWSTrimLargeAddressAwareReserved0040BytesReversedLo32bitMachineDebugStrippedRemovableRunFromSwapNetRunFromSwapSystemDLLUpSystemOnlyBytesReversedHi"

var _Characteristic_map = map[Characteristic]string{
	1:     _Characteristic_name[0:14],
	2:     _Characteristic_name[14:29],
	4:     _Characteristic_name[29:45],
	8:     _Characteristic_name[45:62],
	16:    _Characteristic_name[62:78],
	32:    _Characteristic_name[78:95],
	64:    _Characteristic_name[95:107],
	128:   _Characteristic_name[107:122],
	256:   _Characteristic_name[122:134],
	512:   _Characteristic_name[134:147],
	1024:  _Characteristic_name[147:167],
	2048:  _Characteristic_name[167:181],
	4096:  _Characteristic_name[181:187],
	8192:  _Characteristic_name[187:190],
	16384: _Characteristic_name[190:202],
	32768: _Characteristic_name[202:217],
}

func (i Characteristic) String() string {
	if str, ok := _Characteristic_map[i]; ok {
		return str
	}
	return "Characteristic(" + strconv.FormatInt(int64(i), 10) + ")"
}
