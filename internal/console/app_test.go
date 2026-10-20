package console

import (
	"bytes"
	"strings"
	"testing"

	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

// runScript запускает меню на заданном вводе и возвращает вывод и парк.
func runScript(t *testing.T, lines ...string) (string, *fleet.Fleet) {
	t.Helper()
	f := fleet.New(inspection.NewServiceCenter(inspection.DefaultConfig()),
		fleet.MinSimplicityRule(fleet.DefaultBeginnerThreshold))
	out := &bytes.Buffer{}
	app := New(strings.NewReader(strings.Join(lines, "\n")+"\n"), out, f)
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}
	return out.String(), f
}

func assertContains(t *testing.T, out string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("в выводе нет %q\n---\n%s", w, out)
		}
	}
}

func TestRun_AcceptScooterAndShowEverything(t *testing.T) {
	out, f := runScript(t,
		"1", "1", "Ninebot", "7", "10", "0,45", // самокат, расход с запятой
		"2", "1", "Urban", "M", // шлем
		"3", "4", "5", "6", "0")

	assertContains(t, out,
		"прошёл техосмотр, инв. № VK-T-0001",
		"на балансе, инв. № VK-I-0001",
		"всего на балансе: 2",
		"Суммарное потребление: 0.45",
		"Ninebot",
		"(отказов нет)",
		"До встречи")
	if len(f.Inventory()) != 2 {
		t.Errorf("в парке %d объектов", len(f.Inventory()))
	}
}

func TestRun_InspectionRejectsWornVehicle(t *testing.T) {
	out, f := runScript(t, "1", "2", "Старый велик", "8", "95", "6", "0")

	assertContains(t, out, "отклонён: износ 95%", "Старый велик")
	if len(f.Inventory()) != 0 {
		t.Error("отклонённый транспорт попал в парк")
	}
}

func TestRun_InvalidInputIsReAsked(t *testing.T) {
	out, f := runScript(t,
		"42",          // неизвестный пункт меню
		"1", "9", "3", // вид транспорта: 9 — нет такого, 3 — электровелосипед
		"", "E-Bike", // пустое имя, затем нормальное
		"abc", "11", "4", // простота: не число, вне диапазона, нормально
		"5", "-1", "x", "0.9", // износ 5; расход: отрицательный, не число, нормально
		"0")

	assertContains(t, out,
		"неизвестный пункт меню",
		"нужно целое число от 1 до 3",
		"значение не может быть пустым",
		"нужно целое число от 1 до 10",
		"нужно неотрицательное число",
		"VK-T-0001")
	if got := f.BeginnerVehicles(); len(got) != 0 {
		t.Errorf("простота 4 — не для новичков, а получили %v", got)
	}
}

func TestRun_AllItemTypes(t *testing.T) {
	out, _ := runScript(t,
		"2", "2", "Док", "6", "2.4",
		"2", "3", "Шкаф", "12", "4.8",
		"3", "0")
	assertContains(t, out, "VK-I-0002", "Суммарное потребление: 7.20")
}

func TestRun_EmptyReports(t *testing.T) {
	out, _ := runScript(t, "4", "5", "0")
	assertContains(t, out, "(пока нет)", "(на балансе ничего нет)")
}

func TestRun_EOFInsideDialogEndsGracefully(t *testing.T) {
	out, _ := runScript(t, "1", "1", "Ninebot") // ввод обрывается посреди диалога
	assertContains(t, out, "До встречи")
}

func TestRun_Demo(t *testing.T) {
	out, f := runScript(t, "7", "0")
	assertContains(t, out,
		"VK-T-0003",
		"отклонён: износ 85%",
		"выше нормы",
		"простота для новичка должна быть от 1 до 10",
		"объект уже стоит на балансе",
		"Суммарное потребление: 8.55")
	if r := f.Report(); r.Vehicles != 3 || r.Items != 3 || r.Rejected != 2 {
		t.Errorf("после демо: %+v", r)
	}
}
