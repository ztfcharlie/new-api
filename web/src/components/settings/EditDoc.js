import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Modal,
    Form,
    Checkbox
} from '@douyinfe/semi-ui';

import { API, showError, showSuccess } from '@/helpers';
import Editor from '@/components/settings/Editor'
import {slugify} from '@/helpers/utils'
export default function (props) {
    const { t } = useTranslation();
    const [isEdit, setIsEdit] = useState(false);
    const originInputs = {
        id: undefined,
        title: '',
        slug: '',
        content: '',
        summary: '',
        keywords: '',
        description: '',
        views: 0,
        weight: 0,
        created_at: new Date(),
        updated_at: '',
        status:1,
    };

    const formApi = useRef();
    const [inputs, setInputs] = useState(originInputs);
    const [loading, setLoading] = useState(false);
    const [isSlugReadonly, setIsSlugReadonly] = useState(false);
    const { title,slug, content, summary, keywords, description, views, weight, created_at, status } = inputs
    const handleInputChange = (name, value) => {
        setInputs((inputs) => ({ ...inputs, [name]: value }));
    };
    useEffect(() => {
        setIsEdit(props.editingDoc.id !== undefined);
    }, [props.editingDoc.id]);
    useEffect(() => {
        if (!isEdit) {
            setInputs(originInputs);
        } else {
            loadDoc()
        }
    }, [isEdit]);

    const loadDoc = async () => {
        setLoading(true);
        let res = await API.get(`/api/doc/${props.editingDoc.id}`);
        const { success, message, data } = res.data;
        if (success) {
            data.status = data.status > 0;
            setInputs(data);
            formApi.current.setValues(data);
            // 编辑器的值
            editorRef.current.setContent(data.content)
            setIsSlugReadonly(true)
        } else {
            showError(message);
        }
        setLoading(false);
    };

    const handleCancel = () => {
        setInputs(originInputs);
        setLoading(false);
        props.handleClose();
    };
    async function onSubmit() {
        await formApi.current.validate();
        const values = formApi.current.getValues(); // 获取表单数据
        values.content = inputs.content;
        values.status = values.status ? 1 : -1;
        setLoading(true);
        let res;
        if (isEdit) {
            res = await API.put(`/api/doc/${props.editingDoc.id}`, values);
        } else {
            res = await API.post(`/api/doc`, values);
        }
        const { success, message } = res.data;
        if (success) {
            showSuccess('文档创建成功！');
            setInputs(originInputs);
            props.refresh();
            props.handleClose();
        } else {
            showError(message);
        }
        setLoading(false);
    }
    const editorRef = useRef(null);
    function changeContent(content){
        handleInputChange('content',content);
    }
    return (
        <>
            <Modal
                title={t('添加文章')}
                visible={props.visible}
                confirmLoading={loading}
                onOk={onSubmit}
                onCancel={handleCancel}
                maskClosable={false}
                centered={true}
                fullScreen={true}
                closeOnEsc={true}
                style={{ maxHeight: '100vh' }}
                bodyStyle={{ overflow: 'auto' }}
            >
                <Form labelPosition='top' getFormApi={api => formApi.current = api}>
                    <Form.Input field='title' label='标题' initValue={title}
                        rules={[
                            { required: true, message: '请输入标题' },
                            { min: 3, message: '标题至少 3 个字符' },
                        ]}
                        onChange={v => {
                            handleInputChange('title', v)
                            if(!isSlugReadonly){
                                formApi.current.setValues({...inputs,slug: slugify(v)})
                            }
                        }}
                    ></Form.Input>
                    <Form.Slot label="内容">
                        <Editor ref={editorRef} changeContent={changeContent}/>
                    </Form.Slot>                   
                    <Form.TextArea field='summary' label='内容摘要' initValue={summary}
                        rules={[
                            { required: true, message: '请输入内容摘要' },
                        ]}
                        onChange={v => handleInputChange('summary', v)}
                    ></Form.TextArea>

                    <Form.Input field='slug' label='标识' initValue={slug}
                    {...(isSlugReadonly?{readonly: true}:{})}
                    rules={[
                        { required: true, message: '请输入标识' },
                        { min: 3, message: '标识至少 3 个字符' },
                    ]}
                    onChange={v => handleInputChange('slug', v)}

                    suffix={
                        isEdit?<div class="w-[60px]">
                            <Checkbox checked={isSlugReadonly} onChange={()=>{
                                setIsSlugReadonly(!isSlugReadonly)
                            }}>只读</Checkbox>
                        </div>:null
                        }
                    ></Form.Input>
                                  
                    <Form.Input field='keywords' label='SEO关键词' initValue={keywords}></Form.Input>
                    <Form.TextArea field='description' label='SEO描述' initValue={description}></Form.TextArea>
                    <Form.InputNumber field='views' label='浏览量' initValue={views}></Form.InputNumber>
                    <Form.InputNumber field='weight' label='权重' initValue={weight}></Form.InputNumber>
                    <Form.DatePicker
                        format="yyyy-MM-dd HH:mm:ss"
                        field='created_at'
                        position="top"
                        label={t('添加时间')}
                        style={{ width: 272 }}
                        initValue={created_at}
                        type='dateTime'
                        onChange={(value) => {
                            handleInputChange(value, 'created_at')
                        }}
                    />
                    

                 
                    
                    <Form.Slot label="状态">
                        <div className='flex items-center'>
                            <Form.Switch field='status' label=" "
                            initValue={status>0? true : false}
                            onChange={v => handleInputChange('status', v)}
                            ></Form.Switch>
                            <label className='ml-2'>{inputs.status ? '启用' : '禁用'}</label>
                        </div>
                    </Form.Slot>

                </Form>
            </Modal>
        </>
    );
};

