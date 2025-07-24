package queue

const TAKEOUT = "takeout"
const MEMBER_ORDER_CANCEL = "member_order_cancel"

func Init() {
	initTakeoutCancel()
	initMemberOrderCancel()
}
