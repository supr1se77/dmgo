package utilities

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Masterminds/semver"
)

// A função de verificação de versão foi modificada para exibir uma mensagem local
// em vez de verificar online e exibir mensagens de outro desenvolvedor.
func VersionCheck(version string) {
	LogSuccess("Ferramenta SUPR1SE v%v está pronta.", version)
	LogInfo("Bem-vindo de volta!")
	
}

func Snowflake() int64 {
	snowflake := strconv.FormatInt((time.Now().UTC().UnixNano()/1000000)-1420070400000, 2) + "0000000000000000000000"
	nonce, _ := strconv.ParseInt(snowflake, 2, 64)
	return nonce
}

func ReverseSnowflake(snowflake string) time.Time {
	snowflakei, err := strconv.Atoi(snowflake)
	if err != nil {
		return time.Time{}
	}
	base2 := strconv.FormatInt(int64(snowflakei), 2)
	if len(base2) < 23 {
		return time.Time{}
	}
	ageBase2 := base2[:len(base2)-22]
	t, err := strconv.ParseInt(ageBase2, 2, 64)
	if err != nil {
		return time.Time{}
	}
	t = t + 1420070400000
	tm := time.UnixMilli(t)
	return tm

}

func Contains(s []string, e string) bool {
	defer HandleOutOfBounds()
	if len(s) == 0 {
		return false
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Inputs 2 slices of strings and returns a slice of strings which does not contain elements from the second slice
func RemoveSubset(s []string, r []string) []string {
	var n []string
	for _, v := range s {
		if !Contains(r, v) {
			n = append(n, v)
		}
	}
	return n
}

func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func HandleOutOfBounds() {
	if r := recover(); r != nil {
		fmt.Printf("Recovered from Panic %v", r)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TimeDifference(t1, t2 time.Time) string {
	d := t2.Sub(t1)
	hoursSince := d.Hours()
	years := hoursSince / (24 * 365)
	intYears := int(years)
	remainderYears := years - float64(intYears)
	months := remainderYears * 12
	intMonths := int(months)
	remainderMonths := months - float64(intMonths)
	days := remainderMonths * 30
	intDays := int(days)
	remainderDays := days - float64(intDays)
	hours := remainderDays * 24
	intHours := int(hours)
	return fmt.Sprintf("%v years, %v months, %v days, %v hours", intYears, intMonths, intDays, intHours)
}

func isNewer(check string, current string) bool {
	c, err := semver.NewConstraint(fmt.Sprintf(`>%v`, check))
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	b, _ := c.Validate(v)
	return b
}

func isSame(check string, current string) bool {
	c, err := semver.NewConstraint(fmt.Sprintf(`=%v`, check))
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	b, _ := c.Validate(v)
	return b
}

func isOlder(check string, current string) bool {
	c, err := semver.NewConstraint(fmt.Sprintf(`<%v`, check))
	if err != nil {
		return false
	}
	v, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	b, _ := c.Validate(v)
	return b
}