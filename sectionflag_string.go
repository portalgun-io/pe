// Code generated by "stringer -trimprefix SectionFlag -type SectionFlag"; DO NOT EDIT.

package pe

import "strconv"

const _SectionFlag_name = "Reserved00000000Reserved00000001Reserved00000002Reserved00000004TypeNoPadReserved00000010ContainsCodeContainsInitializedDataContainsUninitializedDataLinkOtherLinkInfoReserved00000400LinkRemoveLinkComdatGPRelMemPurgeableMemLockedMemPreloadAlign1Align2Align4Align8Align16Align32Align64Align128Align256Align512Align1024Align2048Align4096Align8192LinkNRelocOverflowMemDiscardableMemNotCachedMemNotPagedMemSharedMemExecuteMemReadMemWrite"

var _SectionFlag_map = map[SectionFlag]string{
	0:          _SectionFlag_name[0:16],
	1:          _SectionFlag_name[16:32],
	2:          _SectionFlag_name[32:48],
	4:          _SectionFlag_name[48:64],
	8:          _SectionFlag_name[64:73],
	16:         _SectionFlag_name[73:89],
	32:         _SectionFlag_name[89:101],
	64:         _SectionFlag_name[101:124],
	128:        _SectionFlag_name[124:149],
	256:        _SectionFlag_name[149:158],
	512:        _SectionFlag_name[158:166],
	1024:       _SectionFlag_name[166:182],
	2048:       _SectionFlag_name[182:192],
	4096:       _SectionFlag_name[192:202],
	32768:      _SectionFlag_name[202:207],
	131072:     _SectionFlag_name[207:219],
	262144:     _SectionFlag_name[219:228],
	524288:     _SectionFlag_name[228:238],
	1048576:    _SectionFlag_name[238:244],
	2097152:    _SectionFlag_name[244:250],
	3145728:    _SectionFlag_name[250:256],
	4194304:    _SectionFlag_name[256:262],
	5242880:    _SectionFlag_name[262:269],
	6291456:    _SectionFlag_name[269:276],
	7340032:    _SectionFlag_name[276:283],
	8388608:    _SectionFlag_name[283:291],
	9437184:    _SectionFlag_name[291:299],
	10485760:   _SectionFlag_name[299:307],
	11534336:   _SectionFlag_name[307:316],
	12582912:   _SectionFlag_name[316:325],
	13631488:   _SectionFlag_name[325:334],
	14680064:   _SectionFlag_name[334:343],
	16777216:   _SectionFlag_name[343:361],
	33554432:   _SectionFlag_name[361:375],
	67108864:   _SectionFlag_name[375:387],
	134217728:  _SectionFlag_name[387:398],
	268435456:  _SectionFlag_name[398:407],
	536870912:  _SectionFlag_name[407:417],
	1073741824: _SectionFlag_name[417:424],
	2147483648: _SectionFlag_name[424:432],
}

func (i SectionFlag) String() string {
	if str, ok := _SectionFlag_map[i]; ok {
		return str
	}
	return "SectionFlag(" + strconv.FormatInt(int64(i), 10) + ")"
}
