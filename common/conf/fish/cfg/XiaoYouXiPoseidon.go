
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type XiaoYouXiPoseidon struct {
    _dataMap map[int32]*Poseidon
    _dataList []*Poseidon
}

func NewXiaoYouXiPoseidon(_buf []map[string]interface{}) (*XiaoYouXiPoseidon, error) {
    _dataList := make([]*Poseidon, 0, len(_buf))
    dataMap := make(map[int32]*Poseidon)

    for _, _ele_ := range _buf {
        if _v, err2 := NewPoseidon(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Id] = _v
        }
    }
    return &XiaoYouXiPoseidon{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *XiaoYouXiPoseidon) GetDataMap() map[int32]*Poseidon {
    return table._dataMap
}

func (table *XiaoYouXiPoseidon) GetDataList() []*Poseidon {
    return table._dataList
}

func (table *XiaoYouXiPoseidon) Get(key int32) *Poseidon {
    return table._dataMap[key]
}


