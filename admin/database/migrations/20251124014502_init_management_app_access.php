<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitManagementAppAccess extends Migrator
{
    /**
     * 更新或插入数据
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 管理APP权限 
        $shopAccessData = [
            // 管理APP
            ['uuid' => 2856266502144000, 'name' => '管理APP', 'path' => 'management_app', 'api_path' => '', 'parent_uuid' => 0, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 首页
            ['uuid' => 2856287473664000, 'name' => '首页', 'path' => 'home', 'api_path' => '', 'parent_uuid' => 2856266502144000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 店内概况
            ['uuid' => 2856304250880000, 'name' => '店内概况', 'path' => 'store_overview', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 区域数据
            ['uuid' => 2856321028096000, 'name' => '区域数据', 'path' => 'area_data', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 订单数据
            ['uuid' => 2856337805312000, 'name' => '订单数据', 'path' => 'order_data', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 支付数据
            ['uuid' => 2856354582528000, 'name' => '支付数据', 'path' => 'payment_data', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 销量TOP10
            ['uuid' => 2856367165440000, 'name' => '销量TOP10', 'path' => 'sales_top10', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 销售额TOP10
            ['uuid' => 2856388136960000, 'name' => '销售额TOP10', 'path' => 'revenue_top10', 'api_path' => '', 'parent_uuid' => 2856287473664000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 报表中心
            ['uuid' => 2856409108480000, 'name' => '报表中心', 'path' => 'report_center', 'api_path' => '', 'parent_uuid' => 2856266502144000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 营业报表
            ['uuid' => 2856430080000000, 'name' => '营业报表', 'path' => 'business_report', 'api_path' => '', 'parent_uuid' => 2856409108480000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 时段营业统计
            ['uuid' => 2856446857216000, 'name' => '时段营业统计', 'path' => 'period_business_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 综合运营统计
            ['uuid' => 2856476217344000, 'name' => '综合运营统计', 'path' => 'comprehensive_operation_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 营业收款统计
            ['uuid' => 2856505577472000, 'name' => '营业收款统计', 'path' => 'business_payment_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 渠道营业统计
            ['uuid' => 2856543326208000, 'name' => '渠道营业统计', 'path' => 'channel_business_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商品销售统计
            ['uuid' => 2856543326208001, 'name' => '商品销售统计', 'path' => 'product_sales_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 用户分析
            ['uuid' => 2856589463552000, 'name' => '用户分析', 'path' => 'user_analysis', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 门店统计
            ['uuid' => 2856606240768000, 'name' => '门店统计', 'path' => 'store_statistics', 'api_path' => '', 'parent_uuid' => 2856430080000000, 'sort' => 7, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 运营报表
            ['uuid' => 2856623017984000, 'name' => '运营报表', 'path' => 'operation_report', 'api_path' => '', 'parent_uuid' => 2856409108480000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 后厨菜品出品明细
            ['uuid' => 2856635600896000, 'name' => '后厨菜品出品明细', 'path' => 'kitchen_dish_output_details', 'api_path' => '', 'parent_uuid' => 2856623017984000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 后厨效率分析
            ['uuid' => 2856664961024000, 'name' => '后厨效率分析', 'path' => 'kitchen_efficiency_analysis', 'api_path' => '', 'parent_uuid' => 2856623017984000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 导出记录
            ['uuid' => 2856702709760000, 'name' => '导出记录', 'path' => 'export_record', 'api_path' => '', 'parent_uuid' => 2856409108480000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 导出记录
            ['uuid' => 2856723681280000, 'name' => '导出记录', 'path' => 'export_record_list', 'api_path' => '', 'parent_uuid' => 2856702709760000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 工作台
            ['uuid' => 2856757235712000, 'name' => '工作台', 'path' => 'workbench', 'api_path' => '', 'parent_uuid' => 2856266502144000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 初始设置
            ['uuid' => 2856774012928000, 'name' => '初始设置', 'path' => 'initial_setting', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商家档案
            ['uuid' => 2856790790144000, 'name' => '商家档案', 'path' => 'merchant_profile', 'api_path' => '', 'parent_uuid' => 2856774012928000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 员工账号
            ['uuid' => 2856803373056000, 'name' => '员工账号', 'path' => 'staff_account', 'api_path' => '', 'parent_uuid' => 2856774012928000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2856820150272000, 'name' => '添加', 'path' => 'staff_account_add', 'api_path' => '', 'parent_uuid' => 2856803373056000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2856836927488000, 'name' => '编辑', 'path' => 'staff_account_edit', 'api_path' => '', 'parent_uuid' => 2856803373056000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 角色管理
            ['uuid' => 2856866287616000, 'name' => '角色管理', 'path' => 'role_management', 'api_path' => '', 'parent_uuid' => 2856774012928000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2856887259136000, 'name' => '添加', 'path' => 'role_management_add', 'api_path' => '', 'parent_uuid' => 2856866287616000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2856904036352000, 'name' => '编辑', 'path' => 'role_management_edit', 'api_path' => '', 'parent_uuid' => 2856866287616000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2856916619264000, 'name' => '删除', 'path' => 'role_management_delete', 'api_path' => '', 'parent_uuid' => 2856866287616000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 门店管理
            ['uuid' => 2856866287616001, 'name' => '门店管理', 'path' => 'store_management', 'api_path' => '', 'parent_uuid' => 2856774012928000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商品管理
            ['uuid' => 2856933396480000, 'name' => '商品管理', 'path' => 'product_management', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商品管理
            ['uuid' => 2856954368000000, 'name' => '商品管理', 'path' => 'product', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2856971145216000, 'name' => '添加', 'path' => 'product_add', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2856992116736000, 'name' => '编辑', 'path' => 'product_edit', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857013088256000, 'name' => '排序', 'path' => 'product_sort', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857034059776000, 'name' => '删除', 'path' => 'product_delete', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量导入
            ['uuid' => 2857055031296000, 'name' => '批量导入', 'path' => 'product_batch_import', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 5, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量创建Grab
            ['uuid' => 2857076002816000, 'name' => '批量创建Grab', 'path' => 'product_batch_create_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 6, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量上架Grab
            ['uuid' => 2857096974336000, 'name' => '批量上架Grab', 'path' => 'product_batch_online_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 7, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量下架Grab
            ['uuid' => 2857117945856000, 'name' => '批量下架Grab', 'path' => 'product_batch_offline_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 8, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量删除Grab
            ['uuid' => 2857138917376000, 'name' => '批量删除Grab', 'path' => 'product_batch_delete_grab', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 9, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量创建LINE MAN
            ['uuid' => 2857159888896001, 'name' => '批量创建LINE MAN', 'path' => 'product_batch_create_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 10, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量上架LINE MAN
            ['uuid' => 2857180860416001, 'name' => '批量上架LINE MAN', 'path' => 'product_batch_online_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 11, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量下架LINE MAN
            ['uuid' => 2857201831936000, 'name' => '批量下架LINE MAN', 'path' => 'product_batch_offline_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 12, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量删除LINE MAN
            ['uuid' => 2857222803456001, 'name' => '批量删除LINE MAN', 'path' => 'product_batch_delete_lineman', 'api_path' => '', 'parent_uuid' => 2856954368000000, 'sort' => 13, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商品分类
            ['uuid' => 2857080197120000, 'name' => '商品分类', 'path' => 'product_category', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 普通分类
            ['uuid' => 2857105362944000, 'name' => '普通分类', 'path' => 'normal_category', 'api_path' => '', 'parent_uuid' => 2857080197120000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857134723072000, 'name' => '添加', 'path' => 'normal_category_add', 'api_path' => '', 'parent_uuid' => 2857105362944000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857159888896000, 'name' => '编辑', 'path' => 'normal_category_edit', 'api_path' => '', 'parent_uuid' => 2857105362944000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857180860416000, 'name' => '排序', 'path' => 'normal_category_sort', 'api_path' => '', 'parent_uuid' => 2857105362944000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857197637632000, 'name' => '删除', 'path' => 'normal_category_delete', 'api_path' => '', 'parent_uuid' => 2857105362944000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 特色分类
            ['uuid' => 2857222803456000, 'name' => '特色分类', 'path' => 'special_category', 'api_path' => '', 'parent_uuid' => 2857080197120000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857256357888000, 'name' => '添加', 'path' => 'special_category_add', 'api_path' => '', 'parent_uuid' => 2857222803456000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857281523712000, 'name' => '编辑', 'path' => 'special_category_edit', 'api_path' => '', 'parent_uuid' => 2857222803456000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857298300928000, 'name' => '排序', 'path' => 'special_category_sort', 'api_path' => '', 'parent_uuid' => 2857222803456000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857323466752000, 'name' => '删除', 'path' => 'special_category_delete', 'api_path' => '', 'parent_uuid' => 2857222803456000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 规格
            ['uuid' => 2857352826880000, 'name' => '规格', 'path' => 'flavor', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857390575616000, 'name' => '添加', 'path' => 'flavor_add', 'api_path' => '', 'parent_uuid' => 2857352826880000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857411547136000, 'name' => '编辑', 'path' => 'flavor_edit', 'api_path' => '', 'parent_uuid' => 2857352826880000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857432518656000, 'name' => '排序', 'path' => 'flavor_sort', 'api_path' => '', 'parent_uuid' => 2857352826880000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857453490176000, 'name' => '删除', 'path' => 'flavor_delete', 'api_path' => '', 'parent_uuid' => 2857352826880000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 属性
            ['uuid' => 2857487044608000, 'name' => '属性', 'path' => 'attribute', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857533181952000, 'name' => '添加', 'path' => 'attribute_add', 'api_path' => '', 'parent_uuid' => 2857487044608000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857562542080000, 'name' => '编辑', 'path' => 'attribute_edit', 'api_path' => '', 'parent_uuid' => 2857487044608000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857608679424000, 'name' => '排序', 'path' => 'attribute_sort', 'api_path' => '', 'parent_uuid' => 2857487044608000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857633845248000, 'name' => '删除', 'path' => 'attribute_delete', 'api_path' => '', 'parent_uuid' => 2857487044608000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 加料
            ['uuid' => 2857659011072000, 'name' => '加料', 'path' => 'sauce', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857684176896000, 'name' => '添加', 'path' => 'sauce_add', 'api_path' => '', 'parent_uuid' => 2857659011072000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857717731328000, 'name' => '编辑', 'path' => 'sauce_edit', 'api_path' => '', 'parent_uuid' => 2857659011072000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857747091456000, 'name' => '排序', 'path' => 'sauce_sort', 'api_path' => '', 'parent_uuid' => 2857659011072000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857763868672000, 'name' => '删除', 'path' => 'sauce_delete', 'api_path' => '', 'parent_uuid' => 2857659011072000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 单位
            ['uuid' => 2857789034496000, 'name' => '单位', 'path' => 'unit', 'api_path' => '', 'parent_uuid' => 2856933396480000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857810006016000, 'name' => '添加', 'path' => 'unit_add', 'api_path' => '', 'parent_uuid' => 2857789034496000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2857843560448000, 'name' => '编辑', 'path' => 'unit_edit', 'api_path' => '', 'parent_uuid' => 2857789034496000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2857872920576000, 'name' => '排序', 'path' => 'unit_sort', 'api_path' => '', 'parent_uuid' => 2857789034496000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2857893892096000, 'name' => '删除', 'path' => 'unit_delete', 'api_path' => '', 'parent_uuid' => 2857789034496000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 进销存
            ['uuid' => 2857919057920000, 'name' => '进销存', 'path' => 'inventory', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 物品管理
            ['uuid' => 2857952612352000, 'name' => '物品管理', 'path' => 'material_management', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2857986166784000, 'name' => '添加', 'path' => 'material_add', 'api_path' => '', 'parent_uuid' => 2857952612352000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858007138304000, 'name' => '编辑', 'path' => 'material_edit', 'api_path' => '', 'parent_uuid' => 2857952612352000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 菜品导入
            ['uuid' => 2858032304128000, 'name' => '菜品导入', 'path' => 'material_dish_import', 'api_path' => '', 'parent_uuid' => 2857952612352000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量操作
            ['uuid' => 2858061664256000, 'name' => '批量操作', 'path' => 'material_batch_operation', 'api_path' => '', 'parent_uuid' => 2857952612352000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 批量导入
            ['uuid' => 2858086830080000, 'name' => '批量导入', 'path' => 'material_batch_import', 'api_path' => '', 'parent_uuid' => 2857952612352000, 'sort' => 5, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 物品类别
            ['uuid' => 2858103607296000, 'name' => '物品类别', 'path' => 'material_category', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858116190208000, 'name' => '添加', 'path' => 'material_category_add', 'api_path' => '', 'parent_uuid' => 2858103607296000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858128773120000, 'name' => '编辑', 'path' => 'material_category_edit', 'api_path' => '', 'parent_uuid' => 2858103607296000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2858145550336000, 'name' => '排序', 'path' => 'material_category_sort', 'api_path' => '', 'parent_uuid' => 2858103607296000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858162327552000, 'name' => '删除', 'path' => 'material_category_delete', 'api_path' => '', 'parent_uuid' => 2858103607296000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 成本卡
            ['uuid' => 2858179104768000, 'name' => '成本卡', 'path' => 'bom_card', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商品
            ['uuid' => 2858200076288000, 'name' => '商品', 'path' => 'bom_card_product', 'api_path' => '', 'parent_uuid' => 2858179104768000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858233630720000, 'name' => '编辑', 'path' => 'bom_card_product_edit', 'api_path' => '', 'parent_uuid' => 2858200076288000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858216853504000, 'name' => '删除', 'path' => 'bom_card_product_delete', 'api_path' => '', 'parent_uuid' => 2858200076288000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 配置成本卡
            ['uuid' => 2858250407936000, 'name' => '配置成本卡', 'path' => 'bom_card_product_configure', 'api_path' => '', 'parent_uuid' => 2858200076288000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 加料
            ['uuid' => 2858267185152000, 'name' => '加料', 'path' => 'bom_card_sauce', 'api_path' => '', 'parent_uuid' => 2858179104768000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858296545280000, 'name' => '编辑', 'path' => 'bom_card_sauce_edit', 'api_path' => '', 'parent_uuid' => 2858267185152000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858283962368000, 'name' => '删除', 'path' => 'bom_card_sauce_delete', 'api_path' => '', 'parent_uuid' => 2858267185152000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 配置成本卡
            ['uuid' => 2858313322496000, 'name' => '配置成本卡', 'path' => 'bom_card_sauce_configure', 'api_path' => '', 'parent_uuid' => 2858267185152000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 采购申请(外部)
            ['uuid' => 2858330099712000, 'name' => '采购申请(外部)', 'path' => 'purchase_request_external', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858355265536000, 'name' => '添加', 'path' => 'purchase_request_external_add', 'api_path' => '', 'parent_uuid' => 2858330099712000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 审核
            ['uuid' => 2858376237056000, 'name' => '审核', 'path' => 'purchase_request_external_approve', 'api_path' => '', 'parent_uuid' => 2858330099712000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858388819968000, 'name' => '删除', 'path' => 'purchase_request_external_delete', 'api_path' => '', 'parent_uuid' => 2858330099712000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 采购收货(外部)
            ['uuid' => 2858405597184000, 'name' => '采购收货(外部)', 'path' => 'purchase_receipt_external', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 库存查询
            ['uuid' => 2858439151616000, 'name' => '库存查询', 'path' => 'stock_query', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 出入库明细表
            ['uuid' => 2858451734528000, 'name' => '出入库明细表', 'path' => 'in_out_detail', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 7, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 品牌采购(内部)
            ['uuid' => 2858468511744000, 'name' => '品牌采购(内部)', 'path' => 'brand_purchase_internal', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 8, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858493677568000, 'name' => '添加', 'path' => 'brand_purchase_internal_add', 'api_path' => '', 'parent_uuid' => 2858468511744000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 审核
            ['uuid' => 2858518843392000, 'name' => '审核', 'path' => 'brand_purchase_internal_approve', 'api_path' => '', 'parent_uuid' => 2858468511744000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858535620608000, 'name' => '删除', 'path' => 'brand_purchase_internal_delete', 'api_path' => '', 'parent_uuid' => 2858468511744000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 品采收货(内部)
            ['uuid' => 2858548203520000, 'name' => '品采收货(内部)', 'path' => 'brand_purchase_internal_receipt', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 9, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 供应商档案
            ['uuid' => 2858577563648000, 'name' => '供应商档案', 'path' => 'supplier_archive', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 10, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858594340864000, 'name' => '添加', 'path' => 'supplier_archive_add', 'api_path' => '', 'parent_uuid' => 2858577563648000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858611118080000, 'name' => '编辑', 'path' => 'supplier_archive_edit', 'api_path' => '', 'parent_uuid' => 2858577563648000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858636283904000, 'name' => '删除', 'path' => 'supplier_archive_delete', 'api_path' => '', 'parent_uuid' => 2858577563648000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 仓库档案
            ['uuid' => 2858661449728000, 'name' => '仓库档案', 'path' => 'warehouse', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 11, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858678226944000, 'name' => '添加', 'path' => 'warehouse_add', 'api_path' => '', 'parent_uuid' => 2858661449728000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858695004160000, 'name' => '编辑', 'path' => 'warehouse_edit', 'api_path' => '', 'parent_uuid' => 2858661449728000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 设置默认
            ['uuid' => 2858707587072000, 'name' => '设置默认', 'path' => 'warehouse_set_default', 'api_path' => '', 'parent_uuid' => 2858661449728000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858724364288000, 'name' => '删除', 'path' => 'warehouse_delete', 'api_path' => '', 'parent_uuid' => 2858661449728000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 盘点单
            ['uuid' => 2858741141504000, 'name' => '盘点单', 'path' => 'inventory_check', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 12, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858757918720000, 'name' => '添加', 'path' => 'inventory_check_add', 'api_path' => '', 'parent_uuid' => 2858741141504000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 审核
            ['uuid' => 2858787278848000, 'name' => '审核', 'path' => 'inventory_check_approve', 'api_path' => '', 'parent_uuid' => 2858741141504000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858808250368000, 'name' => '删除', 'path' => 'inventory_check_delete', 'api_path' => '', 'parent_uuid' => 2858741141504000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 调拨单
            ['uuid' => 2858825027584000, 'name' => '调拨单', 'path' => 'transfer_order', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 13, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858841804800000, 'name' => '添加', 'path' => 'transfer_order_add', 'api_path' => '', 'parent_uuid' => 2858825027584000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 审核
            ['uuid' => 2858858582016000, 'name' => '审核', 'path' => 'transfer_order_approve', 'api_path' => '', 'parent_uuid' => 2858825027584000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2858875359232000, 'name' => '删除', 'path' => 'transfer_order_delete', 'api_path' => '', 'parent_uuid' => 2858825027584000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 收货
            ['uuid' => 2858892136448000, 'name' => '收货', 'path' => 'transfer_order_receive', 'api_path' => '', 'parent_uuid' => 2858825027584000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 参数设置
            ['uuid' => 2858908913663000, 'name' => '参数设置', 'path' => 'parameter_setting', 'api_path' => '', 'parent_uuid' => 2857919057920000, 'sort' => 14, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 营销管理
            ['uuid' => 2858908913664000, 'name' => '营销管理', 'path' => 'marketing_management', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 菜品营销
            ['uuid' => 2858925690880000, 'name' => '菜品营销', 'path' => 'dish_marketing', 'api_path' => '', 'parent_uuid' => 2858908913664000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2858950856704000, 'name' => '添加', 'path' => 'dish_marketing_add', 'api_path' => '', 'parent_uuid' => 2858925690880000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2858980216832000, 'name' => '编辑', 'path' => 'dish_marketing_edit', 'api_path' => '', 'parent_uuid' => 2858925690880000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 营销活动
            ['uuid' => 2858996994048000, 'name' => '营销活动', 'path' => 'marketing_activity', 'api_path' => '', 'parent_uuid' => 2858908913664000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2859013771264000, 'name' => '添加', 'path' => 'marketing_activity_add', 'api_path' => '', 'parent_uuid' => 2858996994048000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2859030548480000, 'name' => '编辑', 'path' => 'marketing_activity_edit', 'api_path' => '', 'parent_uuid' => 2858996994048000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2859047325696000, 'name' => '删除', 'path' => 'marketing_activity_delete', 'api_path' => '', 'parent_uuid' => 2858996994048000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 外卖管理
            ['uuid' => 2858986936320000, 'name' => '外卖管理', 'path' => 'takeout_management', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // Grab
            ['uuid' => 2859007907840000, 'name' => 'Grab', 'path' => 'takeout_grab', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // LINE MAN
            ['uuid' => 2859028879360000, 'name' => 'LINE MAN', 'path' => 'takeout_lineman', 'api_path' => '', 'parent_uuid' => 2858986936320000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 餐厅设置
            ['uuid' => 2859064102912000, 'name' => '餐厅设置', 'path' => 'restaurant_setting', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 副屏设置
            ['uuid' => 2859080880128000, 'name' => '副屏设置', 'path' => 'secondary_screen_setting', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 票据样式设置
            ['uuid' => 2859106045952000, 'name' => '票据样式设置', 'path' => 'ticket_style_setting', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 桌台地图
            ['uuid' => 2859106045952001, 'name' => '桌台地图', 'path' => 'desk_map', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 打印设置
            ['uuid' => 2859123823104000, 'name' => '打印设置', 'path' => 'printer_setting', 'api_path' => '', 'parent_uuid' => 2859064102912000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 默认票据模板
            ['uuid' => 2859131211776000, 'name' => '默认票据模板', 'path' => 'default_ticket_template', 'api_path' => '', 'parent_uuid' => 2859106045952000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2859147988992000, 'name' => '编辑', 'path' => 'default_ticket_template_edit', 'api_path' => '', 'parent_uuid' => 2859131211776000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 使用
            ['uuid' => 2859164766208000, 'name' => '使用', 'path' => 'default_ticket_template_use', 'api_path' => '', 'parent_uuid' => 2859131211776000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 高级票据模板
            ['uuid' => 2859181543424000, 'name' => '高级票据模板', 'path' => 'advanced_ticket_template', 'api_path' => '', 'parent_uuid' => 2859106045952000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2859198320640000, 'name' => '添加', 'path' => 'advanced_ticket_template_add', 'api_path' => '', 'parent_uuid' => 2859181543424000, 'sort' => 1, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2859210903552000, 'name' => '编辑', 'path' => 'advanced_ticket_template_edit', 'api_path' => '', 'parent_uuid' => 2859181543424000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2859231875072000, 'name' => '删除', 'path' => 'advanced_ticket_template_delete', 'api_path' => '', 'parent_uuid' => 2859181543424000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 使用
            ['uuid' => 2859252846592000, 'name' => '使用', 'path' => 'advanced_ticket_template_use', 'api_path' => '', 'parent_uuid' => 2859181543424000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 其他
            ['uuid' => 2859273818112000, 'name' => '其他', 'path' => 'other', 'api_path' => '', 'parent_uuid' => 2856757235712000, 'sort' => 7, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 各端设置
            ['uuid' => 2859290595328000, 'name' => '各端设置', 'path' => 'terminal_setting', 'api_path' => '', 'parent_uuid' => 2859273818112000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排号取餐
            ['uuid' => 2859311566848000, 'name' => '排号取餐', 'path' => 'queue_meal', 'api_path' => '', 'parent_uuid' => 2859290595328000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 厨显设置
            ['uuid' => 2859332341760000, 'name' => '厨显设置', 'path' => 'kitchen_setting', 'api_path' => '', 'parent_uuid' => 2859290595328000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 自助点餐机
            ['uuid' => 2859353116672000, 'name' => '自助点餐机', 'path' => 'kiosk_setting', 'api_path' => '', 'parent_uuid' => 2859290595328000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 支付管理
            ['uuid' => 2859373891584000, 'name' => '支付管理', 'path' => 'payment_management', 'api_path' => '', 'parent_uuid' => 2859273818112000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 支付配置
            ['uuid' => 2859394666496000, 'name' => '支付配置', 'path' => 'payment_config', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 添加
            ['uuid' => 2859415441408000, 'name' => '添加', 'path' => 'payment_add', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 2, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 编辑
            ['uuid' => 2859436216320000, 'name' => '编辑', 'path' => 'payment_edit', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 3, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 排序
            ['uuid' => 2859456991232000, 'name' => '排序', 'path' => 'payment_sort', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 4, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 删除
            ['uuid' => 2859477766144000, 'name' => '删除', 'path' => 'payment_delete', 'api_path' => '', 'parent_uuid' => 2859373891584000, 'sort' => 5, 'is_route' => 0, 'is_menu' => 0, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 我的
            ['uuid' => 2859328344064000, 'name' => '我的', 'path' => 'my', 'api_path' => '', 'parent_uuid' => 2856266502144000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 业务设置
            ['uuid' => 2859345121280000, 'name' => '业务设置', 'path' => 'business_setting', 'api_path' => '', 'parent_uuid' => 2859328344064000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 优惠设置
            ['uuid' => 2859361898496000, 'name' => '优惠设置', 'path' => 'discount_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 1, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 优惠计入设置
            ['uuid' => 2859387064320000, 'name' => '优惠计入设置', 'path' => 'discount_count_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 二维码
            ['uuid' => 2859412230144000, 'name' => '二维码', 'path' => 'qr_code', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 3, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 基础设置
            ['uuid' => 2859433201664000, 'name' => '基础设置', 'path' => 'basic_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 4, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 敏感操作设置
            ['uuid' => 2859441589952000, 'name' => '敏感操作设置', 'path' => 'profile_business_setting_sensitive', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 5, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 原因管理
            ['uuid' => 2859449978880000, 'name' => '原因管理', 'path' => 'reason_management', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 6, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 分批送厨规则
            ['uuid' => 2859466756096000, 'name' => '分批送厨规则', 'path' => 'batch_cooking_rule', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 7, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 安全库存设置
            ['uuid' => 2859479339008000, 'name' => '安全库存设置', 'path' => 'safety_stock_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 8, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 点餐设置
            ['uuid' => 2859479339008001, 'name' => '点餐设置', 'path' => 'take_meal_setting', 'api_path' => '', 'parent_uuid' => 2859345121280000, 'sort' => 9, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
            // 商家档案
            ['uuid' => 2859496116224000, 'name' => '商家档案', 'path' => 'merchant_profile_my', 'api_path' => '', 'parent_uuid' => 2859328344064000, 'sort' => 2, 'is_route' => 1, 'is_menu' => 1, 'is_show' => 1, 'is_supplier' => 0, 'create_time' => 1763721369, 'update_time' => 1763721369],
        ];
        $this->updateOrInsertData('access', 'uuid', $shopAccessData);
        // 店长角色
        $managers = $db->name('role')->where('id', '>', '0')->where('delete_time', '=', '0')->select();
        foreach ($managers as $manager) {
            if ($manager && isset($manager['uuid'])) {
                $managerRoleData = [
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856266502144000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856287473664000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856304250880000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856321028096000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856337805312000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856354582528000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856367165440000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856388136960000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856409108480000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856430080000000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856446857216000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856476217344000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856505577472000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856543326208000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856543326208001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856589463552000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856606240768000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856623017984000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856635600896000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856664961024000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856702709760000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856723681280000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856757235712000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856774012928000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856790790144000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856803373056000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856820150272000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856836927488000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856866287616000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856887259136000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856904036352000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856916619264000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856866287616001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856933396480000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856954368000000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856971145216000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2856992116736000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857013088256000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857034059776000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857055031296000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857076002816000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857096974336000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857117945856000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857138917376000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857159888896001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857180860416001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857201831936000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857222803456001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857080197120000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857105362944000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857134723072000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857159888896000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857180860416000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857197637632000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857222803456000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857256357888000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857281523712000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857298300928000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857323466752000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857352826880000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857390575616000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857411547136000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857432518656000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857453490176000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857487044608000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857533181952000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857562542080000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857608679424000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857633845248000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857659011072000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857684176896000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857717731328000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857747091456000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857763868672000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857789034496000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857810006016000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857843560448000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857872920576000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857893892096000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857919057920000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857952612352000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2857986166784000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858007138304000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858032304128000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858061664256000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858086830080000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858103607296000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858116190208000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858128773120000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858145550336000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858162327552000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858179104768000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858200076288000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858216853504000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858233630720000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858250407936000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858267185152000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858283962368000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858296545280000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858313322496000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858330099712000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858355265536000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858376237056000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858388819968000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858405597184000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858439151616000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858451734528000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858468511744000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858493677568000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858518843392000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858535620608000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858548203520000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858577563648000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858594340864000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858611118080000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858636283904000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858661449728000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858678226944000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858695004160000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858707587072000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858724364288000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858741141504000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858757918720000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858787278848000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858808250368000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858825027584000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858841804800000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858858582016000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858875359232000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858892136448000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858908913663000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858908913664000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858925690880000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858950856704000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858980216832000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858996994048000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859013771264000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859030548480000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859047325696000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2858986936320000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859007907840000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859028879360000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859064102912000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859080880128000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859106045952000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859106045952001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859123823104000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859131211776000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859147988992000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859164766208000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859181543424000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859198320640000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859210903552000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859231875072000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859252846592000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859273818112000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859290595328000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859311566848000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859332341760000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859353116672000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859373891584000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859394666496000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859415441408000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859436216320000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859456991232000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859477766144000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859328344064000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859345121280000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859361898496000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859387064320000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859412230144000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859433201664000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859441589952000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859449978880000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859466756096000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859479339008000', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859479339008001', 'create_time' => time()],
                    ['uuid' => createUuid(), 'role_uuid' => $manager['uuid'], 'access_uuid' => '2859496116224000', 'create_time' => time()],
                ];
                $this->updateOrInsertData('role_access', ['role_uuid', 'access_uuid'], $managerRoleData);
            }
        }
    }

    /**
     * @param string $tableName 表名
     * @param string|array $uniqueKey 唯一键
     * @param array $data 数据
     */
    private function updateOrInsertData($tableName, $uniqueKey, $data)
    {
        $db = Db::connect(Db::getConfig('default'), true);
        //
        foreach ($data as $item) {
            $query = $db->name($tableName);
            if (is_array($uniqueKey)) {
                foreach ($uniqueKey as $key) {
                    $query->where($key, '=', $item[$key]);
                }
            } else {
                $query->where($uniqueKey, '=', $item[$uniqueKey]);
            }

            $existingData = $query->find();
            if ($existingData) {
                // $query->update($item);
            } else {
                $db->name($tableName)->insert($item);
            }
        }
    }
}
