
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


import "errors"

type BufferZhuanPanConfig struct {
    CellNum int32
    Prob []float32
    Rate []int32
}

const TypeId_BufferZhuanPanConfig = 1200185259

func (*BufferZhuanPanConfig) GetTypeId() int32 {
    return 1200185259
}

func NewBufferZhuanPanConfig(_buf map[string]interface{}) (_v *BufferZhuanPanConfig, err error) {
    _v = &BufferZhuanPanConfig{}
    { var _ok_ bool; var _tempNum_ float64; if _tempNum_, _ok_ = _buf["cellNum"].(float64); !_ok_ { err = errors.New("cellNum error"); return }; _v.CellNum = int32(_tempNum_) }
     {
                    var _arr_ []interface{}
                    var _ok_ bool
                    if _arr_, _ok_ = _buf["prob"].([]interface{}); !_ok_ { err = errors.New("prob error"); return }
    
                    _v.Prob = make([]float32, 0, len(_arr_))
                    
                    for _, _e_ := range _arr_ {
                        var _list_v_ float32
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _e_.(float64); !_ok_ { err = errors.New("_list_v_ error"); return }; _list_v_ = float32(_x_) }
                        _v.Prob = append(_v.Prob, _list_v_)
                    }
                }

     {
                    var _arr_ []interface{}
                    var _ok_ bool
                    if _arr_, _ok_ = _buf["rate"].([]interface{}); !_ok_ { err = errors.New("rate error"); return }
    
                    _v.Rate = make([]int32, 0, len(_arr_))
                    
                    for _, _e_ := range _arr_ {
                        var _list_v_ int32
                        { var _ok_ bool; var _x_ float64; if _x_, _ok_ = _e_.(float64); !_ok_ { err = errors.New("_list_v_ error"); return }; _list_v_ = int32(_x_) }
                        _v.Rate = append(_v.Rate, _list_v_)
                    }
                }

    return
}

