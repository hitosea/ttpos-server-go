<?php
use think\migration\Migrator;

class UpdateLinemanName extends Migrator
{
    /**
     * 将外卖来源 "Line Man" 更新为 "LINE MAN"
     * 任务 #38257: 优化-新管理端-业务设置-外卖，默认存在的Line Man改成 LINE MAN
     */
    public function change()
    {
        // 更新多语言名称表中的 Line Man 为 LINE MAN
        $this->execute("
            UPDATE ttpos_multi_language_name
            SET zh_name = 'LINE MAN',
                en_name = 'LINE MAN',
                zh_tw_name = 'LINE MAN',
                th_name = 'LINE MAN',
                my_name = 'LINE MAN',
                ja_name = 'LINE MAN',
                ko_name = 'LINE MAN',
                tr_name = 'LINE MAN',
                sv_name = 'LINE MAN',
                update_time = UNIX_TIMESTAMP()
            WHERE zh_name = 'Line Man'
               OR en_name = 'Line Man'
        ");
    }
}
