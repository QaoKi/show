
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type RobotConfig struct {
    RoomType int32
    Coin *TwoIntBean
    ChangeGunCoin []*ChangeGunBean
}

const TypeId_RobotConfig = -622491092

func (*RobotConfig) GetTypeId() int32 {
    return -622491092
}

func NewRobotConfig(_buf map[string]interface{}) (_v *RobotConfig, err error) {
    _v = &RobotConfig{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["roomType"].(float64); !_ok_ { err = errors.New("roomType error"); return }; _v.RoomType = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["coin"].(map[string]interface{}); !_ok_ { err = errors.New("coin error"); return }; if _v.Coin, err = NewTwoIntBean(_x_); err != nil { return } }
     {
                    var _arr_ []interface{}
                    var _ok_ bool
                    if _arr_, _ok_ = _buf["changeGunCoin"].([]interface{}); !_ok_ { err = errors.New("changeGunCoin error"); return }
    
                    _v.ChangeGunCoin = make([]*ChangeGunBean, 0, len(_arr_))
                    
                    for _, _e_ := range _arr_ {
                        var _list_v_ *ChangeGunBean
                        { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _e_.(map[string]interface{}); !_ok_ { err = errors.New("_list_v_ error"); return }; if _list_v_, err = NewChangeGunBean(_x_); err != nil { return } }
                        _v.ChangeGunCoin = append(_v.ChangeGunCoin, _list_v_)
                    }
                }

    return
}

