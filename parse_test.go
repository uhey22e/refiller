package refiller

import (
	"bytes"
	"encoding/csv"
	"io"
	"testing"
	"time"
)

func TestParseCsv(t *testing.T) {
	data := bytes.NewReader([]byte(`id,name
1,"Marty McFly"
2,"Emmett Brown"`))

	type testRow struct {
		ID   int       `scan:"id" column:"id"`
		Name string    `scan:"name" column:"name"`
		Date time.Time `column:"date"`
	}

	cr := csv.NewReader(data)
	header, err := cr.Read()
	if err == io.EOF {
		t.Fatal("empty csv")
		return
	} else if err != nil {
		t.Fatal(err)
	}

	base := &testRow{
		Date: time.Date(2015, 10, 21, 0, 0, 0, 0, time.Local),
	}
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		buf := *base
		if err = ParseRow(&buf, header, row); err != nil {
			t.Fatal(err)
		}
		t.Log(Values(buf))
	}
}

func TestParseRow(t *testing.T) {
	type args struct {
		dest   interface{}
		header []string
		row    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				dest: &(struct {
					Name  string    `scan:"name"`
					Score int       `scan:"score" scanDefault:"100"`
					Power float32   `scan:"数値"`
					Date  time.Time `scan:"date" scanTimeFormat:"2006-01-02"`
					Time  time.Time `scan:"time" scanTimeFormat:"15:04"`
				}{}),
				header: []string{"name", "数値", "date", "time"},
				row:    []string{"Marty McFly", "123.4", "2015-10-21", "12:34"},
			},
		},
		{
			args: args{
				dest: &(struct {
					Name string `scan:"name"`
				}{}),
				header: []string{},
				row:    []string{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseRow(tt.args.dest, tt.args.header, tt.args.row); (err != nil) != tt.wantErr {
				t.Errorf("ParseRow() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Log(tt.args.dest)
		})
	}
}
