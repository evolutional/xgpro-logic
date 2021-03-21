package xgpro

type byte lgcFileVoltageLevel

const (
	lgcMaxItemCount = 512,

	lgcFileVoltageLevel5v0 = 0,
	lgcFileVoltageLevel3v3 = 1,
	lgcFileVoltageLevel2v5 = 2,
	lgcFileVoltageLevel1v8 = 3 
)


type struct lgcFileHeader {
	uint32 allCrc32
	uint32 uiFlag
	uint32 itemCount
	uint32 res
	uint32 []itemStart
}

type struct lgcFileItem {
	uint32 vectorCount
	byte itemName
	lgcFileVoltageLevel voltage
	byte pinCount
	byte res0
	byte res1
	uint uiRes	 
}

func ParseLogicFile(fileName string) {
	
}