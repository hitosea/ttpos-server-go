package setting

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"ttpos-server-go/app/dto/resp/setting"
)

func TestName(t *testing.T) {

	dbStr := `{"carousel":[{"file_path":"\/api\/product\/thumbs\/uploads\/shop1724054084\/20250122\/3b7a95d73b3b6d138808855eb06df010.png","real_name":"image.png","sort":"0","type":"image"}],"is_auto_send":"0","order_method":{"is_cashier_order":"1","is_table_order":"1"},"server":{"ip":"192.168.97.4","port":"8080"},"is_remain_color":"1","remain_color":["#E50028","#F2A000"],"advanced_password":"666888","is_open_cashier_password":"1","cashier_password":"666888","lock_password":"666888","is_auto_lock_screen":"1","auto_lock_screen":"300","is_show_scan_sold_out":1,"is_show_assistant_sold_out":1,"default_language":"th","is_auto_order":"0","auto_order_limit":"1000","is_auto_voice":"0","menu_show_sold_out":"1","kitchen_language":"th","language":["th","zh","zhtw","en","ja"]}`

	fmt.Println(strings.Contains(dbStr, "\"is_show_scan_sold_out\""))

	var dddd setting.Cashier
	e := json.Unmarshal([]byte(dbStr), &dddd)
	if e != nil {
		panic(e)
	}

	fmt.Println(dddd.IsShowScanSoldOut)
	fmt.Println("+++++")
}
