
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type Athena struct {
    Id int32
    Rate *TwoIntBean
    Count *TwoIntBean
}

const TypeId_Athena = 1971222083

func (*Athena) GetTypeId() int32 {
    return 1971222083
}

func NewAthena(_buf map[string]interface{}) (_v *Athena, err error) {
    _v = &Athena{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["id"].(float64); !_ok_ { err = errors.New("id error"); return }; _v.Id = int32(_tempNum_) }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["rate"].(map[string]interface{}); !_ok_ { err = errors.New("rate error"); return }; if _v.Rate, err = NewTwoIntBean(_x_); err != nil { return } }
    { var _ok_ bool; var _x_ map[string]interface{}; if _x_, _ok_ = _buf["count"].(map[string]interface{}); !_ok_ { err = errors.New("count error"); return }; if _v.Count, err = NewTwoIntBean(_x_); err != nil { return } }
    return
}
