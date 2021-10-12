package mongo_test

import (
	"fmt"
	"github.com/yuyitech/db"
	"github.com/yuyitech/db/adapter/mongo"
	"reflect"
	"testing"
)

func TestQueryFilter(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want string
	}{
		// string
		{
			name: `db.Cond{"EmailAddress *=": "foo"}`,
			args: db.Cond{"EmailAddress *=": "foo"},
			want: `[{EmailAddress {"pattern": "^foo", "options": "gim"}}]`,
		},
		{
			name: `db.Cond{"EmailAddress =*": "foo"}`,
			args: db.Cond{"EmailAddress =*": "foo"},
			want: `[{EmailAddress {"pattern": "foo$", "options": "gim"}}]`,
		},
		{
			name: `db.Cond{"EmailAddress *": "foo"}`,
			args: db.Cond{"EmailAddress *": "foo"},
			want: `[{EmailAddress {"pattern": "foo", "options": "gim"}}]`,
		},
		// number
		{
			name: `db.Cond{"Status": 1}`,
			args: db.Cond{"Status": 1},
			want: `[{Status 1}]`,
		},
		{
			name: `db.Cond{"Status=": 1}`,
			args: db.Cond{"Status": 1},
			want: `[{Status 1}]`,
		},
		{
			name: `db.Cond{"Status !=": 1}`,
			args: db.Cond{"Status !=": 1},
			want: `[{Status {$ne 1}}]`,
		},
		{
			name: `db.Cond{"Status >": 0}`,
			args: db.Cond{"Status >": 0},
			want: `[{Status {$gt 0}}]`,
		},
		{
			name: `db.Cond{"Status >=": 1}`,
			args: db.Cond{"Status >=": 1},
			want: `[{Status {$gte 1}}]`,
		},
		{
			name: `db.Cond{"Status <": 0}`,
			args: db.Cond{"Status <": 0},
			want: `[{Status {$lt 0}}]`,
		},
		{
			name: `db.Cond{"Status <=": 0}`,
			args: db.Cond{"Status <=": 0},
			want: `[{Status {$lte 0}}]`,
		},
		// range
		{
			name: `db.Cond{"Status $in": []int{1, -1, -2}}`,
			args: db.Cond{"Status $in": []int{1, -1, -2}},
			want: `[{Status {$in [1 -1 -2]}}]`,
		},
		{
			name: `db.Cond{"Status $nin": []int{-1, -2}}`,
			args: db.Cond{"Status $nin": []int{-1, -2}},
			want: `[{Status {$nin [-1 -2]}}]`,
		},
		{
			name: `db.And(db.Cond{"CreatedAt >=": 1633536000}, db.Cond{"CreatedAt <=": 1633622399})`,
			args: db.And(
				db.Cond{"CreatedAt >=": 1633536000},
				db.Cond{"CreatedAt <=": 1633622399},
			),
			want: `[{$and [[{CreatedAt {$gte 1633536000}}] [{CreatedAt {$lte 1633622399}}]]}]`,
		},
		{
			name: `db.And(db.Or(db.Cond{"Username": "foo"}, db.Cond{"Username": "bar"}), db.Cond{"Status": 1})`,
			args: db.And(
				db.Or(
					db.Cond{"Username": "foo"},
					db.Cond{"Username": "bar"},
				),
				db.Cond{"Status": 1},
			),
			want: `[{$and [[{$or [[{Username foo}] [{Username bar}]]}] [{Status 1}]]}]`,
		},
		// exists
		{
			name: `db.Cond{"PhoneNumber $exists": true}`,
			args: db.Cond{"PhoneNumber $exists": true},
			want: `[{PhoneNumber {$exists true}}]`,
		},
		{
			name: `db.Cond{"PhoneNumber $exists": false}`,
			args: db.Cond{"PhoneNumber $exists": false},
			want: `[{PhoneNumber {$exists false}}]`,
		},
		// logic
		{
			name: `db.And(db.Cond{"Username": "foo"}, db.Cond{"Username": "bar"})`,
			args: db.And(
				db.Cond{"Username": "foo"},
				db.Cond{"Username": "bar"},
			),
			want: `[{$and [[{Username foo}] [{Username bar}]]}]`,
		},
		{
			name: `db.Or(db.And(db.Cond{"CountryCode": "86"}, db.Cond{"PhoneNumber": "13800138000"}), db.Cond{"EmailAddress $exists": true})`,
			args: db.Or(
				db.And(
					db.Cond{"CountryCode": "86"},
					db.Cond{"PhoneNumber": "13800138000"},
				),
				db.Cond{"EmailAddress $exists": true},
			),
			want: `[{$or [[{$and [[{CountryCode 86}] [{PhoneNumber 13800138000}]]}] [{EmailAddress {$exists true}}]]}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mongo.QueryFilter(tt.args); !reflect.DeepEqual(fmt.Sprintf("%v", got), tt.want) {
				t.Errorf("QueryFilter() = %v, want %v", got, tt.want)
			}
		})
	}

}
