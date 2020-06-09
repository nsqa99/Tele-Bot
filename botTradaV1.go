package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Step int

const (
	stepToChooseCourse Step = iota
	stepToEnterDifferentCousre
	stepDone
)

type Display struct {
	name     string
	course   string
	extend string
	yearob   string
	dif      int
}

type User struct {
	name              string
	course            string
	phonenum          string
	yearob            string
	lastSigninRequest time.Time
	registrationStep  Step
}


var userMap = make(map[int]*User)
var u User
var ud Display

// var courseMap = map[int]string{
// 	111: "Dapp",
// 	112: "Machine Learning",
// 	113: "Web Development",
// }

func main() {
	b, err := tb.NewBot(tb.Settings{

		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 2 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tb.OnText, func(m *tb.Message) {

		checkStep(b, m)
	})

	// b.Handle("/me", func(m *tb.Message) {
	// 	b.Send(m.Sender, "Tính năng đang được phát triển, vui lòng chọn tính năng khác trong mục /help")
	// })

	b.Handle("/start", func(m *tb.Message) {
		if m.FromGroup() {
			b.Send(m.Sender, "Xin hãy chat riêng với bot để đăng kí.")
		} else {

			b.Send(m.Sender, "Cảm ơn bạn đã liên lạc với Trada Tech. Đây là bot trả lời tự động hỗ trợ các lệnh sau:\n/info - thông tin về dịch vụ và khoá học\n/register - đăng kí khoá học, dịch vụ, hoặc đặt câu hỏi cho Trada Tech.\n/cancel - thông báo muốn huỷ đăng kí.\n/help - hiện hướng dẫn.")
		}
	})
	b.Handle("/help", func(m *tb.Message) {

		b.Send(m.Sender, "/info - thông tin về dịch vụ và khoá học\n/register - đăng kí khoá học, dịch vụ, hoặc đặt câu hỏi cho Trada Tech.\n/cancel - thông báo muốn huỷ đăng kí.\n/help - hiện hướng dẫn.")

	})
	b.Handle("/info", func(m *tb.Message) {

		b.Send(m.Sender, "Thông tin về dịch vụ và khoá học:\n1. Đào tạo Ethereum DApp Developer\n2. Đào tạo theo yêu cầu riêng của công ty bạn\n3. Tư vấn về phát triển và ứng dụng blockchain")

	})

	b.Handle("/register", func(m *tb.Message) {
		if m.FromGroup() {
			b.Send(m.Sender, "Xin hãy chat riêng với bot để đăng kí.")
		} else {
			startRegistration(b, m)
			//next(b, m)
		}

	})

	b.Handle("/cancel", func(m *tb.Message) {

		if ud.course == "" {
			sendAndHideKeyboard(b, m, "Bạn vẫn chưa đăng kí khoá học nào, không thể huỷ khoá học!")
		} else {
			sendAndHideKeyboard(b, m, "Yêu cầu huỷ của bạn đã được thông báo cho nhân viên Trada Tech. Chúng tôi sẽ liên lạc lại với bạn để xác nhận.")
			sendCancelRequest(b, m)
		}

	})

	b.Start()
}

func containAny(array []string, item string) bool {
	for _, element := range array {
		if strings.EqualFold(element, item) {
			return true
		}
	}

	return false
}

func isCourse1(text string) bool {
	values := []string{"DApp"}
	//, "course 111", "111", "course 1", "first course"
	return containAny(values, strings.TrimSpace(text))
}

func isCourse2(text string) bool {
	values := []string{"Khác", "Khac"}
	//, "course 112", "112", "course 2", "second course"
	return containAny(values, strings.TrimSpace(text))
}

func listCourse(b *tb.Bot, m *tb.Message) {
	//sendfAndHideKeyboard(b, m, "Các khoá học hiện có:\nDApp - Khoá học Ethereum DApp Development\nKhác - Các yêu cầu khác")

	sendCourseChoices(b, m, "Các khoá học hiện có:\nDApp - Khoá học Ethereum DApp Development\nKhác - Các yêu cầu khác\nVui lòng chọn khoá học:")
}

func confirmDisplayYear(b *tb.Bot, m *tb.Message) {
	sendYesNo(b, m, "Bạn có muốn nhập và hiển thị công khai năm sinh không?")
}

func sendCourseChoices(b *tb.Bot, m *tb.Message, text string) (*tb.Message, error) {

	c1Btn := tb.ReplyButton{Text: "DApp"}
	c2Btn := tb.ReplyButton{Text: "Khác"}
	//c3Btn := tb.ReplyButton{Text: "Web Development"}
	replyChoice := [][]tb.ReplyButton{
		[]tb.ReplyButton{c1Btn, c2Btn},
		// []tb.ReplyButton{c2Btn},
		//[]tb.ReplyButton{c3Btn},
	}

	return b.Send(m.Sender,
		text,
		&tb.ReplyMarkup{
			ReplyKeyboard:       replyChoice,
			ResizeReplyKeyboard: true,
			OneTimeKeyboard:     true,
			ReplyKeyboardRemove: true,
		})
}

func sendYesNo(b *tb.Bot, m *tb.Message, text string) (*tb.Message, error) {
	yesBtn := tb.ReplyButton{Text: "Yes"}
	noBtn := tb.ReplyButton{Text: "No"}
	replyYesNo := [][]tb.ReplyButton{
		[]tb.ReplyButton{yesBtn, noBtn},
	}
	return b.Send(m.Sender,
		text,
		&tb.ReplyMarkup{
			ReplyKeyboard:       replyYesNo,
			ResizeReplyKeyboard: true,
			OneTimeKeyboard:     true,
		})
}


func sendMessageToAdmin(b *tb.Bot, m *tb.Message) {
	var text string
	username := m.Sender.Username
	//sendfAndHideKeyboard(b, m, "Hello @%s!", m.Sender.Username)

	if username == "" {
		url := "[*người dùng private*](tg://user?id="+strconv.Itoa(m.Sender.ID)+")"
		//text = "tg://user?id=" + strconv.Itoa(m.Sender.ID)

		if ud.dif == 0 {
			text = "Có đăng kí mới!%0ATừ: " + url + "%0AKhoá học: " + ud.course + "%0AThông tin thêm:"
		} else {
			text = "Có đăng kí mới!%0ATừ: " + url + "%0AKhoá học: " + ud.course + "%0AThông tin thêm: " + ud.extend
		}
		text+="&parse_mode=markdown"
		
	
	} else {
		//text = "Hey boss, @" + username + " has just registered!%0AHere's the info:%0ATelegram ID: " + strconv.Itoa(m.Sender.ID) + "%0ADisplay name: " + ud.name + "%0ACourse: " + ud.course + "%0APhone number: " + ud.phonenum + "%0AYear of birth: " + ud.yearob

		if ud.dif == 0 {
			text = "Có đăng kí mới!%0ATừ: @" + username + "%0AKhoá học: " + ud.course + "%0AThông tin thêm:"
		} else {
			text = "Có đăng kí mới!%0ATừ: @" + username + "%0AKhoá học: " + ud.course + "%0AThông tin thêm: " + ud.extend
		}

	}

	_, err := http.Get("https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage?chat_id=" + os.Getenv("_ID") + "&text=" + text)
		if err != nil {
			fmt.Print("error: %s", err)
		}

}



func sendCancelRequest(b *tb.Bot, m *tb.Message) {
	var text string
	username := m.Sender.Username
	//sendfAndHideKeyboard(b, m, "Hello @%s!", m.Sender.Username)

	if username == "" {
		//text = "Hey boss, someone without an username has just registered!%0AHere's the info:%0ATelegram ID: " + strconv.Itoa(m.Sender.ID) + "%0AName: " + ud.name + "%0ACourse: " + ud.course + "%0APhone number: " + ud.phonenum + "%0AYear of birth: " + ud.yearob
		text = "Người dùng private vừa huỷ khoá học mới đăng kí."
	} else {
		//text = "Hey boss, @" + username + " has just registered!%0AHere's the info:%0ATelegram ID: " + strconv.Itoa(m.Sender.ID) + "%0ADisplay name: " + ud.name + "%0ACourse: " + ud.course + "%0APhone number: " + ud.phonenum + "%0AYear of birth: " + ud.yearob
		text = "@" + username + " vừa huỷ khoá học mới đăng kí."

	}

	_, err := http.Get("https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/sendMessage?chat_id=" + os.Getenv("_ID") + "&text=" + text)
	if err != nil {
		fmt.Print("error: %s", err)
	}
}

func startRegistration(b *tb.Bot, m *tb.Message) {
	newUser := User{registrationStep: stepToChooseCourse}
	userMap[m.Sender.ID] = &newUser
	listCourse(b, m)
	//askDisplayName(b, m)

}

func sendAndHideKeyboard(b *tb.Bot, m *tb.Message, text string) (*tb.Message, error) {
	return b.Send(m.Sender, text, &tb.ReplyMarkup{ReplyKeyboardRemove: true})
}
func sendfAndHideKeyboard(b *tb.Bot, m *tb.Message, text string, a ...interface{}) (*tb.Message, error) {
	return sendAndHideKeyboard(b, m, fmt.Sprintf(text, a...))
}

func sayGoodBye(b *tb.Bot, m *tb.Message) {

	sendAndHideKeyboard(b, m, "\nCảm ơn bạn đã đăng kí, thông tin của bạn đã được gửi cho nhân viên Trada Tech xử lý, chúng tôi sẽ liên lạc lại để xác nhận thông tin.")
}

func differentCourse(b *tb.Bot, m *tb.Message) {
	sendAndHideKeyboard(b, m, "Vui lòng cung cấp thêm chi tiết về khoá học bạn muốn đăng kí")
}

func awaitCommand(b *tb.Bot, m *tb.Message) {
	b.Send(m.Sender, "\nGõ /help để hiển thị menu trợ giúp.")
}

func checkStep(b *tb.Bot, m *tb.Message) {

	if u, ok := userMap[m.Sender.ID]; ok {
		switch u.registrationStep {
		case stepToChooseCourse:
			if isCourse1(m.Text) {
				u.registrationStep = stepDone
				ud.dif = 0
				u.course = strings.Title(strings.TrimSpace(m.Text))
				ud.course = u.course

				sayGoodBye(b, m)
				sendMessageToAdmin(b, m)
				next(b, m)
			}
			if isCourse2(m.Text) {
				u.registrationStep = stepToEnterDifferentCousre
				u.course = strings.Title(strings.TrimSpace(m.Text))
				ud.course = u.course
				ud.dif = 1
				//sendAndHideKeyboard(b,m,"Cảm ơn bạn. Nhân viên của Trada Tech sẽ liên lạc lại với bạn để hỏi thêm chi tiết.")
				next(b, m)
			} else if !isCourse1(m.Text) && !isCourse2(m.Text) {
				sendAndHideKeyboard(b, m, "Vui lòng dùng 2 nút có sẵn để trả lời.")

				next(b, m)
			}
			
		case stepToEnterDifferentCousre:
			u.registrationStep = stepDone
			//u.course = strings.Title(strings.TrimSpace(m.Text))
			ud.extend = strings.Title(strings.TrimSpace(m.Text))
			sendAndHideKeyboard(b, m, "Nhân viên của Trada Tech sẽ liên lạc lại với bạn để hỏi thêm chi tiết, cảm ơn bạn đã đăng kí.")
			sendMessageToAdmin(b, m)
			//	sayGoodBye(b, m)
			next(b, m)
			//removeRegisteredUser(m)

		default:
			//u.registrationStep = stepToChooseCourse
			awaitCommand(b, m)

		}

	} else {

		awaitCommand(b, m)
		//sayGoodBye(b, m)
		//sendMessageToAdmin()
	}
}

func next(b *tb.Bot, m *tb.Message) {

	if user, ok := userMap[m.Sender.ID]; ok {
		funcArray := []func(*tb.Bot, *tb.Message){
			listCourse,
			differentCourse,
			removeRegisteredUser,
		}
		funcArray[user.registrationStep](b, m)
	} else {

		startRegistration(b, m)
	}
}

func removeRegisteredUser(b *tb.Bot, m *tb.Message) {
	//b.Send(m.Sender, "")
	delete(userMap, m.Sender.ID)
}
