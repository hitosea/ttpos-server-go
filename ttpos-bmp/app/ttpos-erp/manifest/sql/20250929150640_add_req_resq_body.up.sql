ALTER TABLE erp_receive_cancel_pos_invoice ADD req_body TEXT NULL COMMENT '请求文本，如果能转换';
ALTER TABLE erp_receive_cancel_pos_invoice ADD resp_body TEXT NULL COMMENT '响应文本，如果能转换';



ALTER TABLE erp_receive_close_pos ADD req_body TEXT NULL COMMENT '请求文本，如果能转换';
ALTER TABLE erp_receive_close_pos ADD resp_body TEXT NULL COMMENT '响应文本，如果能转换';


ALTER TABLE erp_receive_pos_invoice ADD req_body TEXT NULL COMMENT '请求文本，如果能转换';
ALTER TABLE erp_receive_pos_invoice ADD resp_body TEXT NULL COMMENT '响应文本，如果能转换';


ALTER TABLE erp_receive_return_pos_invoice ADD req_body TEXT NULL COMMENT '请求文本，如果能转换';
ALTER TABLE erp_receive_return_pos_invoice ADD resp_body TEXT NULL COMMENT '响应文本，如果能转换';
