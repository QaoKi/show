
//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------

package cfg;


type TableSpecialFormat3 struct {
    _dataMap map[int32]*SpecialFormat3
    _dataList []*SpecialFormat3
}

func NewTableSpecialFormat3(_buf []map[string]interface{}) (*TableSpecialFormat3, error) {
    _dataList := make([]*SpecialFormat3, 0, len(_buf))
    dataMap := make(map[int32]*SpecialFormat3)

    for _, _ele_ := range _buf {
        if _v, err2 := NewSpecialFormat3(_ele_); err2 != nil {
            return nil, err2
        } else {
            _dataList = append(_dataList, _v)
            dataMap[_v.ScreenDirection] = _v
        }
    }
    return &TableSpecialFormat3{_dataList:_dataList, _dataMap:dataMap}, nil
}

func (table *TableSpecialFormat3) GetDataMap() map[int32]*SpecialFormat3 {
    return table._dataMap
}

func (table *TableSpecialFormat3) GetDataList() []*SpecialFormat3 {
    return table._dataList
}

func (table *TableSpecialFormat3) Get(key int32) *SpecialFormat3 {
    return table._dataMap[key]
}


