
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type TableSpecialFormat1 struct {
    _dataMap map[int32]*SpecialFormat1
    _dataList []*SpecialFormat1
}

func NewTableSpecialFormat1(_buf []map[string]interface{}) (*TableSpecialFormat1, error) {
    _dataList := make([]*SpecialFormat1, 0, len(_buf))
    dataMap := make(map[int32]*SpecialFormat1)

    for _, _ele_ := range _buf {
        if _v, err2 := NewSpecialFormat1(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.Lot] = _v
        }
    }
    return &TableSpecialFormat1{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *TableSpecialFormat1) GetDataMap() map[int32]*SpecialFormat1 {
    return table._dataMap
}

func (table *TableSpecialFormat1) GetDataList() []*SpecialFormat1 {
    return table._dataList
}

func (table *TableSpecialFormat1) Get(key int32) *SpecialFormat1 {
    return table._dataMap[key]
}


