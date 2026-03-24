package dump

import (
	"testing"
)

var testNames = []string{
	"Документ.РеализацияТоваров.МодульОбъекта",
	"Документ.РеализацияТоваров.МодульМенеджера",
	"Документ.ПоступлениеТоваров.МодульОбъекта",
	"Документ.ПоступлениеТоваров.МодульМенеджера",
	"Документ.ПоступлениеТоваров.Форма.ФормаДокумента.МодульФормы",
	"Справочник.Номенклатура.МодульОбъекта",
	"Справочник.Номенклатура.МодульМенеджера",
	"Справочник.Контрагенты.МодульОбъекта",
	"ОбщийМодуль.ОбщегоНазначения.Модуль",
	"ОбщийМодуль.РаботаСФайлами.Модуль",
	"РегистрСведений.КурсыВалют.МодульНабораЗаписей",
}

func TestNewPathIndex(t *testing.T) {
	pi := NewPathIndex(testNames)

	if pi.Count() != len(testNames) {
		t.Errorf("Count() = %d, want %d", pi.Count(), len(testNames))
	}

	// Verify categories are indexed.
	cats := pi.Categories()
	if len(cats) != 4 {
		t.Errorf("Categories() = %v, want 4 categories", cats)
	}

	// Verify all entries are present.
	for _, name := range testNames {
		if !pi.Contains(name) {
			t.Errorf("Contains(%q) = false, want true", name)
		}
	}
}

func TestNewPathIndexEmpty(t *testing.T) {
	pi := NewPathIndex(nil)
	if pi.Count() != 0 {
		t.Errorf("Count() = %d, want 0", pi.Count())
	}
	if cats := pi.Categories(); len(cats) != 0 {
		t.Errorf("Categories() = %v, want empty", cats)
	}
}

func TestNewPathIndexNil(t *testing.T) {
	var pi *PathIndex
	if pi.Count() != 0 {
		t.Errorf("nil PathIndex Count() = %d, want 0", pi.Count())
	}
	if cats := pi.Categories(); cats != nil {
		t.Errorf("nil PathIndex Categories() = %v, want nil", cats)
	}
	if entries := pi.Filter("", "", ""); entries != nil {
		t.Errorf("nil PathIndex Filter() = %v, want nil", entries)
	}
}

func TestPathIndexFilterByCategory(t *testing.T) {
	pi := NewPathIndex(testNames)

	entries := pi.Filter("Документ", "", "")
	if len(entries) != 5 {
		t.Errorf("Filter(Документ) = %d entries, want 5", len(entries))
	}
	for _, e := range entries {
		if e.Category != "Документ" {
			t.Errorf("entry %q has category %q, want Документ", e.DocID, e.Category)
		}
	}
}

func TestPathIndexFilterByModuleType(t *testing.T) {
	pi := NewPathIndex(testNames)

	entries := pi.Filter("", "", "МодульОбъекта")
	if len(entries) != 4 {
		t.Errorf("Filter(МодульОбъекта) = %d entries, want 4", len(entries))
	}
	for _, e := range entries {
		if e.ModuleType != "МодульОбъекта" {
			t.Errorf("entry %q has moduleType %q, want МодульОбъекта", e.DocID, e.ModuleType)
		}
	}
}

func TestPathIndexFilterByObjectName(t *testing.T) {
	pi := NewPathIndex(testNames)

	entries := pi.Filter("", "Номенклатура", "")
	if len(entries) != 2 {
		t.Errorf("Filter(Номенклатура) = %d entries, want 2", len(entries))
	}
	for _, e := range entries {
		if e.ObjectName != "Номенклатура" {
			t.Errorf("entry %q has objectName %q, want Номенклатура", e.DocID, e.ObjectName)
		}
	}
}

func TestPathIndexFilterCombined(t *testing.T) {
	pi := NewPathIndex(testNames)

	// Category + ModuleType
	entries := pi.Filter("Документ", "", "МодульОбъекта")
	if len(entries) != 2 {
		t.Errorf("Filter(Документ, МодульОбъекта) = %d entries, want 2", len(entries))
	}

	// Category + ObjectName
	entries = pi.Filter("Документ", "РеализацияТоваров", "")
	if len(entries) != 2 {
		t.Errorf("Filter(Документ, РеализацияТоваров) = %d entries, want 2", len(entries))
	}

	// All three filters
	entries = pi.Filter("Документ", "РеализацияТоваров", "МодульОбъекта")
	if len(entries) != 1 {
		t.Errorf("Filter(Документ, РеализацияТоваров, МодульОбъекта) = %d entries, want 1", len(entries))
	}
	if len(entries) == 1 && entries[0].DocID != "Документ.РеализацияТоваров.МодульОбъекта" {
		t.Errorf("expected Документ.РеализацияТоваров.МодульОбъекта, got %q", entries[0].DocID)
	}
}

func TestPathIndexFilterNoMatch(t *testing.T) {
	pi := NewPathIndex(testNames)

	entries := pi.Filter("НесуществующаяКатегория", "", "")
	if len(entries) != 0 {
		t.Errorf("Filter(НесуществующаяКатегория) = %d entries, want 0", len(entries))
	}
}

func TestPathIndexFilterDocIDs(t *testing.T) {
	pi := NewPathIndex(testNames)

	ids := pi.FilterDocIDs("Справочник", "")
	if len(ids) != 3 {
		t.Errorf("FilterDocIDs(Справочник) = %d, want 3", len(ids))
	}

	ids = pi.FilterDocIDs("Справочник", "МодульОбъекта")
	if len(ids) != 2 {
		t.Errorf("FilterDocIDs(Справочник, МодульОбъекта) = %d, want 2", len(ids))
	}
}

func TestPathIndexCategories(t *testing.T) {
	pi := NewPathIndex(testNames)

	cats := pi.Categories()
	expected := []string{"Документ", "ОбщийМодуль", "РегистрСведений", "Справочник"}
	if len(cats) != len(expected) {
		t.Fatalf("Categories() = %v, want %v", cats, expected)
	}
	for i, c := range cats {
		if c != expected[i] {
			t.Errorf("Categories()[%d] = %q, want %q", i, c, expected[i])
		}
	}
}

func TestPathIndexObjects(t *testing.T) {
	pi := NewPathIndex(testNames)

	objs := pi.Objects("Документ")
	expected := []string{"ПоступлениеТоваров", "РеализацияТоваров"}
	if len(objs) != len(expected) {
		t.Fatalf("Objects(Документ) = %v, want %v", objs, expected)
	}
	for i, o := range objs {
		if o != expected[i] {
			t.Errorf("Objects(Документ)[%d] = %q, want %q", i, o, expected[i])
		}
	}
}

func TestPathIndexObjectsAll(t *testing.T) {
	pi := NewPathIndex(testNames)

	objs := pi.Objects("")
	if len(objs) != 7 {
		t.Errorf("Objects('') = %d, want 7 unique object names", len(objs))
	}
}

func TestPathIndexModuleTypes(t *testing.T) {
	pi := NewPathIndex(testNames)

	types := pi.ModuleTypes("Документ", "ПоступлениеТоваров")
	expected := []string{"МодульМенеджера", "МодульОбъекта", "МодульФормы"}
	if len(types) != len(expected) {
		t.Fatalf("ModuleTypes(Документ, ПоступлениеТоваров) = %v, want %v", types, expected)
	}
	for i, mt := range types {
		if mt != expected[i] {
			t.Errorf("ModuleTypes()[%d] = %q, want %q", i, mt, expected[i])
		}
	}
}

func TestPathIndexCount(t *testing.T) {
	pi := NewPathIndex(testNames)
	if pi.Count() != 11 {
		t.Errorf("Count() = %d, want 11", pi.Count())
	}
}

func TestPathIndexFormModuleParsing(t *testing.T) {
	pi := NewPathIndex([]string{
		"Документ.ПоступлениеТоваров.Форма.ФормаДокумента.МодульФормы",
	})

	entries := pi.Filter("Документ", "", "МодульФормы")
	if len(entries) != 1 {
		t.Fatalf("Filter for form module = %d entries, want 1", len(entries))
	}
	e := entries[0]
	if e.Category != "Документ" {
		t.Errorf("Category = %q, want Документ", e.Category)
	}
	if e.ObjectName != "ПоступлениеТоваров" {
		t.Errorf("ObjectName = %q, want ПоступлениеТоваров", e.ObjectName)
	}
	if e.ModuleType != "МодульФормы" {
		t.Errorf("ModuleType = %q, want МодульФормы", e.ModuleType)
	}
}

func BenchmarkPathIndexFilter(b *testing.B) {
	// Generate a realistic set of names.
	categories := []string{"Документ", "Справочник", "Обработка", "Отчет", "РегистрСведений"}
	moduleTypes := []string{"МодульОбъекта", "МодульМенеджера", "МодульФормы"}
	var names []string
	for _, cat := range categories {
		for i := range 1000 {
			for _, mt := range moduleTypes {
				name := cat + ".Объект" + string(rune('A'+i%26)) + "." + mt
				names = append(names, name)
			}
		}
	}
	pi := NewPathIndex(names)

	b.ResetTimer()
	for range b.N {
		pi.Filter("Документ", "", "МодульОбъекта")
	}
}
