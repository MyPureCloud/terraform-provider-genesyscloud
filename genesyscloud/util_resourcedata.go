package genesyscloud

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func setNillableValue[T any](d *schema.ResourceData, key string, value *T) {
	if value != nil {
		d.Set(key, *value)
	} else {
		d.Set(key, nil)
	}
}

func getNillableValue[T any](d *schema.ResourceData, key string) *T {
	value, ok := d.GetOk(key)
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

// More info about using deprecated GetOkExists: https://github.com/hashicorp/terraform-plugin-sdk/issues/817
func getNillableBool(d *schema.ResourceData, key string) *bool {
	value, ok := d.GetOkExists(key)
	if ok {
		v := value.(bool)
		return &v
	}
	return nil
}
