package trix

import (
	"testing"
)

var (
	sampleDefren = []string{
		"de.2=zwei",
		"de.1=eins",
		"fr.1=un",
		"fr.2=deux",
		"fr.3=trois",
		"en.1=one",
		"en.2=two",
		"en.3=three",
		"en.4=four",
		"en.4=five",
	}
	sampleTypes = []string{
		"val.int1=1",
		"val.int2=-1",
		"val.bool.true.1=1",
		"val.bool.true.2=t",
		"val.bool.true.3=true",
		"val.bool.true.4=on",
		"val.bool.true.5=T",
		"val.bool.true.6=TRUE",
		"val.bool.true.7=ON",
		"val.bool.false.1=0",
		"val.bool.false.2=f",
		"val.bool.false.3=false",
		"val.bool.false.4=off",
		"val.bool.false.5=F",
		"val.bool.false.6=FALSE",
		"val.bool.false.7=OFF",
		"val.bool.false.8=",
	}
	sampleSett = []string{
		// 1 key, one and multiple values
		"main.settings.types.1.keys.1=category",
		"main.settings.types.1.1020.value=s,u,k",
		"main.settings.types.1.8020.value=s",
		"main.settings.types.2.default=s,k",
		// 2 keys, gaps on numbering, star
		"main.settings.params.4.keys.1=category",
		"main.settings.params.4.keys.2=type",
		"main.settings.params.4.2021.s.value=a,b,c",
		"main.settings.params.4.2022.s.value=a,b,c,d",
		"main.settings.params.4.2023.*.value=star",
		"main.settings.params.9.default=a",
		// continue
		"main.settings.data.1.default=first",
		"main.settings.data.1.continue=1",
		"main.settings.data.2.keys.1=category",
		"main.settings.data.2.2021.value=second",
		// no default
		"main.settings.stuff.1.keys.1=category",
		"main.settings.stuff.1.2021.value=x,y",
	}
)

func TestSettings(t *testing.T) {
	root := NewRoot()
	root.SetKey(`settings.types.1.keys.1`, `category`)
	root.SetKey(`settings.types.1.1001.value`, `sell,rent,buy`)
	root.SetKey(`settings.types.1.1002.value`, `sell,rent,buy,donation`)
	root.SetKey(`settings.types.1.1003.value`, `rent,buy`)
	root.SetKey(`settings.types.2.default`, `sell`)

	root.SetKey(`settings.params.1.keys.1`, `category`)
	root.SetKey(`settings.params.1.keys.2`, `type`)
	root.SetKey(`settings.params.1.1001.sell.value`, `price`)
	root.SetKey(`settings.params.1.1002.*.value`, `price,mileage`)
	root.SetKey(`settings.params.1.continue`, `1`)
	root.SetKey(`settings.params.2.default`, `color`)

	root.SetKey(`settings.images.1.keys.1`, `?category`)
	root.SetKey(`settings.images.1.false.value`, `max:0`)
	root.SetKey(`settings.images.2.keys.1`, `?type`)
	root.SetKey(`settings.images.2.false.value`, `max:0`)
	root.SetKey(`settings.images.3.keys.1`, `type`)
	root.SetKey(`settings.images.3.buy.value`, `max:0`)
	root.SetKey(`settings.images.4.keys.1`, `category`)
	root.SetKey(`settings.images.4.1001.value`, `max:12,extra:4,extra_price:5`)
	root.SetKey(`settings.images.4.1002.value`, `max:12`)
	root.SetKey(`settings.images.4.1003.value`, `max:0,comment:Easy as 1\,2\,3`)
	root.SetKey(`settings.images.5.default`, `max:8`)
	root.SortRecursively()

	c := func(lastKey string, added Args, expected Reply) {
		t.Helper()
		testDeepEqual(t, root.With(added).GetSettings("settings", lastKey), expected)
	}

	// 1-level keys, default
	c("types", Args{}, Reply{"value": {"sell"}})
	c("types", Args{"category": 1001}, Reply{"value": {"sell", "rent", "buy"}})
	c("types", Args{"category": 1002}, Reply{"value": {"sell", "rent", "buy", "donation"}})
	c("types", Args{"category": 1003}, Reply{"value": {"rent", "buy"}})
	c("types", Args{"category": 1099}, Reply{"value": {"sell"}}) // default

	// 2-level keys, star-keys continue, default
	c("params", Args{}, Reply{"value": {"color"}})
	c("params", Args{"category": 1001}, Reply{"value": {"color"}})
	c("params", Args{"category": "1001"}, Reply{"value": {"color"}})
	c("params", Args{"type": "sell"}, Reply{"value": {"color"}})
	c("params", Args{"category": 1001, "type": "sell"}, Reply{"value": {"price", "color"}})
	c("params", Args{"category": 1002, "type": "sell"}, Reply{"value": {"price", "mileage", "color"}})
	c("params", Args{"category": 1002, "type": "whatever"}, Reply{"value": {"price", "mileage", "color"}})

	// ?keys, named values, escaping
	c("images", Args{}, Reply{"max": {"0"}})
	c("images", Args{"category": 1001}, Reply{"max": {"0"}}) // no type: max:0
	c("images", Args{"type": "sell"}, Reply{"max": {"0"}})   // no category: max:0
	c("images", Args{"category": 1099, "type": "whatever"}, Reply{"max": {"8"}})
	c("images", Args{"category": 1001, "type": "whatever"}, Reply{"max": {"12"}, "extra": {"4"}, "extra_price": {"5"}})
	c("images", Args{"category": 1003, "type": "whatever"}, Reply{"max": {"0"}, "comment": {"Easy as 1,2,3"}})
}
