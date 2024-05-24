package tfexporter

var DataSourceExports []string

func SetDataSourceExports() []string {
	if len(DataSourceExports) < 1 {
		DataSourceExports = append(DataSourceExports, "")
	}
	return DataSourceExports
}
