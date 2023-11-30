
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type RobotAction struct {
    Id int32
    GameTime *TwoIntBean
    ChangeGunTime *TwoIntBean
    ChangeActionTime *TwoIntBean
    RobotActionList []int32
    SuoDingTime *TwoIntBean
    SuoDingFishId int32
    ChangeAngleTime *TwoIntBean
    GunSpeedTime *TwoIntBean
    GunSpeed *TwoIntBean
    DrawTime *TwoIntBean
    PoseidonTime *TwoIntBean
    HitPorb *TwoFloatBean
    HitPorbTime *TwoIntBean
}

const TypeId_RobotAction = -690649824

func (*RobotAction) GetTypeId() int32 {
    return -690649824
}

func NewRobotAction(_buf map[string]interface{}) (_v *RobotAction, err error) {
    _v = &RobotAction{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["gameTime"].(map[string]interface{}); !_ok_ { err = errors.New("gameTime error"); return }; if _v.GameTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["changeGunTime"].(map[string]interface{}); !_ok_ { err = errors.New("changeGunTime error"); return }; if _v.ChangeGunTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["changeActionTime"].(map[string]interface{}); !_ok_ { err = errors.New("changeActionTime error"); return }; if _v.ChangeActionTime, err = NewTwoIntBean(_x_); err != nil { return } }
     {
                    var _arr_ []interface{}
                    var _ok_ bool
                    if _arr_, _ok_ = _buf["robotActionList"].([]interface{}); !_ok_ { err = errors.New("robotActionList error"); return }
    
                    _v.RobotActionList = make([]int32, 0, len(_arr_))
                    
                    for _, _e_ := range _arr_ {
                        var _list_v_ int32
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _e_.(float64); !_ok_ { err = errors.New("_list_v_ error"); return }; _list_v_ = int32(_x_) }
                        _v.RobotActionList = append(_v.RobotActionList, _list_v_)
                    }
                }

    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["suoDingTime"].(map[string]interface{}); !_ok_ { err = errors.New("suoDingTime error"); return }; if _v.SuoDingTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["suoDingFishId"].(float64); !_ok_ { err = errors.New("suoDingFishId error"); return }; _v.SuoDingFishId = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["changeAngleTime"].(map[string]interface{}); !_ok_ { err = errors.New("changeAngleTime error"); return }; if _v.ChangeAngleTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["gunSpeedTime"].(map[string]interface{}); !_ok_ { err = errors.New("gunSpeedTime error"); return }; if _v.GunSpeedTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["gunSpeed"].(map[string]interface{}); !_ok_ { err = errors.New("gunSpeed error"); return }; if _v.GunSpeed, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["drawTime"].(map[string]interface{}); !_ok_ { err = errors.New("drawTime error"); return }; if _v.DrawTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["poseidonTime"].(map[string]interface{}); !_ok_ { err = errors.New("poseidonTime error"); return }; if _v.PoseidonTime, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["hitPorb"].(map[string]interface{}); !_ok_ { err = errors.New("hitPorb error"); return }; if _v.HitPorb, err = NewTwoFloatBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["hitPorbTime"].(map[string]interface{}); !_ok_ { err = errors.New("hitPorbTime error"); return }; if _v.HitPorbTime, err = NewTwoIntBean(_x_); err != nil { return } }
    return
}

