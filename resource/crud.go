package resource

import (
	"errors"
	"fmt"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"

	"github.com/Sky-And-Hammer/TM_EC"
	"github.com/Sky-And-Hammer/TM_EC/utils"
	"github.com/Sky-And-Hammer/roles"
)

func (res *Resource) findOneHandler(result interface{}, metaValues *MetaValues, context *TM_EC.Context) error {
	if res.HasPermission(roles.Read, context) {
		var (
			scope        = context.GetDB().NewScope(res.Value)
			primaryField = res.PrimaryField()
			primaryKey   string
		)

		if metaValues == nil {
			primaryKey = context.ResourceID
		} else if primaryField == nil {
			return nil
		} else if id := metaValues.Get(primaryField.Name); id != nil {
			primaryKey = utils.ToString(id.Value)
		}

		if primaryKey != "" {
			if metaValues != nil {
				if destory := metaValues.Get("_destory"); destory != nil {
					if fmt.Sprint(destory.Value) != "0" && res.HasPermission(roles.Delete, context) {
						context.GetDB().Delete(result, fmt.Sprintf("%v = ?", scope.Quote(primaryField.DBName)), primaryKey)
						return ErrProcessorSkipLeft
					}
				}
			}
			return context.GetDB().First(result, fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(primaryField.DBName)), primaryKey).Error
		}
		return errors.New("failed to find")
	}
	return roles.ErrPermissionDenied
}

func (res *Resource) findManyHandler(result interface{}, context *TM_EC.Context) error {
	if res.HasPermission(roles.Read, context) {
		db := context.GetDB()
		if _, ok := db.Get("ec:getting_total_count"); ok {
			return context.GetDB().Count(result).Error
		} else {
			return context.GetDB().Set("gorn:order_by_primary_key", "DESC").Find(result).Error
		}
	}
	return roles.ErrPermissionDenied
}

func (res *Resource) saveHandler(result interface{}, context *TM_EC.Context) error {
	if context.GetDB().NewScope(result).PrimaryKeyZero() &&
		res.HasPermission(roles.Create, context) ||
		res.HasPermission(roles.Update, context) {
		return context.GetDB().Save(result).Error
	}

	return roles.ErrPermissionDenied
}

func (res *Resource) deleteHandler(result interface{}, context *TM_EC.Context) error {
	if res.HasPermission(roles.Delete, context) {
		scope := context.GetDB().NewScope(res.Value)
		if !context.GetDB().First(result, fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName())), context.ResourceID).RecordNotFound() {
			return context.GetDB().Delete(result).Error
		}
		return gorm.ErrRecordNotFound
	}
	return roles.ErrPermissionDenied
}

func (res *Resource) CallFindOne(result interface{}, metaValues *MetaValues, context *TM_EC.Context) error {
	return res.FindOneHandler(result, metaValues, context)
}

func (res *Resource) CallFindMany(result interface{}, context *TM_EC.Context) error {
	return res.FindManyHandler(result, context)
}

func (res *Resource) CallSave(result interface{}, context *TM_EC.Context) error {
	return res.SaveHandler(result, context)
}

func (res *Resource) CallDelete(result interface{}, context *TM_EC.Context) error {
	return res.DeleteHandler(result, context)
}
