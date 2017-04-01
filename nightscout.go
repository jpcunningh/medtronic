package medtronic

import (
	"log"
	"time"

	"github.com/ecc1/nightscout"
)

func Treatments(records []HistoryRecord) []nightscout.Treatment {
	var treatments []nightscout.Treatment
	user := nightscout.Username()
	insulin0 := Insulin(0).NightscoutInsulin()
	duration0 := 0
	for i, r := range records {
		info := nightscout.Treatment{
			CreatedAt: time.Time(r.Time),
			EnteredBy: user,
		}
		var r2 *HistoryRecord
		if i+1 < len(records) {
			r2 = &records[i+1]
		}
		switch r.Type() {
		case BGCapture:
			info.EventType = "BG Check"
			g := r.Glucose.NightscoutGlucose()
			info.Glucose = &g
			info.Units = "mg/dl"
		case TempBasalRate:
			if !nextEvent(r, r2, TempBasalDuration) {
				continue
			}
			info.EventType = "Temp Basal"
			if *r2.Duration == 0 {
				info.Absolute = &insulin0
				info.Duration = &duration0
			} else {
				ins := r.Insulin.NightscoutInsulin()
				info.Absolute = &ins
				min := int(*r2.Duration / Duration(time.Minute))
				info.Duration = &min
			}
		case Bolus:
			info.EventType = "Meal Bolus"
			ins := r.Bolus.Amount.NightscoutInsulin()
			info.Insulin = &ins
			min := int(r.Bolus.Duration / Duration(time.Minute))
			info.Duration = &min
		case Rewind:
			if !nextEvent(r, r2, Prime) {
				continue
			}
			info.EventType = "Site Change"
		case ResumePump:
			info.EventType = "Temp Basal"
			info.Absolute = &insulin0
			info.Duration = &duration0
		case SuspendPump:
			info.EventType = "Temp Basal"
			info.Absolute = &insulin0
			min := 24 * 60
			info.Duration = &min
		default:
			continue
		}
		treatments = append(treatments, info)
	}
	return treatments
}

func nextEvent(r HistoryRecord, r2 *HistoryRecord, t HistoryRecordType) bool {
	if r2 == nil {
		log.Printf("expected %v to be followed by %v at %v", r.Type(), t, r.Time)
		return false
	}
	if r2.Type() != t {
		log.Printf("expected %v to be followed by %v at %v but found %v", r.Type(), t, r.Time, r2.Type())
		return false
	}
	return true
}

func (r Glucose) NightscoutGlucose() nightscout.Glucose {
	return nightscout.Glucose(r)
}

func (r Insulin) NightscoutInsulin() nightscout.Insulin {
	return nightscout.Insulin(float64(r) / 1000)
}

func (r Voltage) NightscoutVoltage() nightscout.Voltage {
	return nightscout.Voltage(float64(r) / 1000)
}

func (sched BasalRateSchedule) NightscoutSchedule() nightscout.Schedule {
	n := len(sched)
	tv := make(nightscout.Schedule, n)
	for i, r := range sched {
		tv[i] = nightscout.TimeValue{
			Time:  r.Start.String(),
			Value: r.Rate,
		}
	}
	return tv
}

func (sched CarbRatioSchedule) NightscoutSchedule() nightscout.Schedule {
	n := len(sched)
	if n != 0 && sched[0].Units != Grams {
		panic("carb units must be grams")
	}
	tv := make(nightscout.Schedule, n)
	for i, r := range sched {
		tv[i] = nightscout.TimeValue{
			Time:  r.Start.String(),
			Value: float64(r.Ratio) / 10, // Grams
		}
	}
	return tv
}

func (sched InsulinSensitivitySchedule) NightscoutSchedule() nightscout.Schedule {
	n := len(sched)
	tv := make(nightscout.Schedule, n)
	for i, r := range sched {
		tv[i] = nightscout.TimeValue{
			Time:  r.Start.String(),
			Value: r.Sensitivity,
		}
	}
	return tv
}

func (sched GlucoseTargetSchedule) NightscoutSchedule() (nightscout.Schedule, nightscout.Schedule) {
	n := len(sched)
	low := make(nightscout.Schedule, n)
	high := make(nightscout.Schedule, n)
	for i, r := range sched {
		t := r.Start.String()
		low[i] = nightscout.TimeValue{
			Time:  t,
			Value: r.Low,
		}
		high[i] = nightscout.TimeValue{
			Time:  t,
			Value: r.High,
		}
	}
	return low, high
}
