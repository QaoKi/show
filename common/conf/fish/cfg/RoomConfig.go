
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type RoomConfig struct {
    RoomType int32
    Name string
    InitTableCount int32
    MinCoin int32
    MinExp int32
    MinRate int32
    MaxRate int32
}

const TypeId_RoomConfig = -1837073379

func (*RoomConfig) GetTypeId() int32 {
    return -1837073379
}

func NewRoomConfig(_buf map[string]interface{}) (_v *RoomConfig, err error) {
    _v = &RoomConfig{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["roomType"].(float64); !_ok_ { err = errors.New("roomType error"); return }; _v.RoomType = int32(_tempNum_) }
    { var _ok_ bool; if _v.Name, _ok_ = _buf["name"].(string); !_ok_ { err = errors.New("name error"); return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["initTableCount"].(float64); !_ok_ { err = errors.New("initTableCount error"); return }; _v.InitTableCount = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["minCoin"].(float64); !_ok_ { err = errors.New("minCoin error"); return }; _v.MinCoin = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["minExp"].(float64); !_ok_ { err = errors.New("minExp error"); return }; _v.MinExp = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["minRate"].(float64); !_ok_ { err = errors.New("minRate error"); return }; _v.MinRate = int32(_tempNum_) }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["maxRate"].(float64); !_ok_ { err = errors.New("maxRate error"); return }; _v.MaxRate = int32(_tempNum_) }
    return
}

