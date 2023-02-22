package data

import (
	"clothing-store/internal/validator"
	"strings"
)

type Filters struct {
	Page         int64
	PageSize     int64
	Sort         string
	SortSafelist []string
}

type Keys struct {
	PriceMax       int64
	PriceMin       int64
	Brand          string
	Sizes          []string
	SizesSafelist  []string
	BrandsSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	// Check that the sort parameter matches a value in the safelist.
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func ValidateKeys(v *validator.Validator, k Keys) {

	for i := 0; i < len(k.BrandsSafelist); i++ {
		k.BrandsSafelist[i] = strings.ToLower(k.BrandsSafelist[i])
	}
	v.Check(k.PriceMin >= 0, "price_min", "must be greater or equal to zero")
	v.Check(k.PriceMax > 0, "price_max", "must be greater than zero")
	v.Check(k.PriceMax > k.PriceMin, "price", "price_max must be greater than price_min")
	v.Check(validator.PermittedValue(strings.ToLower(k.Brand), k.BrandsSafelist...), "brand", "do not have this brand")
	for i := 0; i < len(k.Sizes); i++ {
		v.Check(validator.PermittedValue(strings.ToUpper(k.Sizes[i]), k.SizesSafelist...), "size", "invalid size value")
	}
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int64 {
	return f.PageSize
}
func (f Filters) offset() int64 {
	return (f.Page - 1) * f.PageSize
}
