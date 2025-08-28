import React, { useMemo, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
    API,
    showError,
    showSuccess,
} from '@/helpers';

import {
    Button,
    Form,
    Layout,
    Table,
    Tag,
    Space,
} from '@douyinfe/semi-ui';
import { ITEMS_PER_PAGE } from '@/constants';
import EditDoc from './EditDoc'
import { formatDateTime } from '@/helpers/datetime';
import { Popconfirm } from '@douyinfe/semi-ui';


const DocsTable = () => {
    const { t } = useTranslation();
    const [editingDoc, setEditingDoc] = useState({
        id: undefined,
    });
    const [docList, setDocList] = useState([]);
    const [loading, setLoading] = useState(false);
    const [docCount, setDocCount] = useState(ITEMS_PER_PAGE);
    const [activePage, setActivePage] = useState(1);
    const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
    const [inputs, setInputs] = useState({
        id: '',
        title: '',
    });
    const [showEdit, setShowEdit] = useState(false)
    const columns = [
        { title: t('Id'), dataIndex: 'id', width: 80 },
        { title: t('标题'), dataIndex: 'title', width: 300, ellipsis: true },
        /*{ title: '内容摘要', dataIndex: 'summary', width: 200, ellipsis: true },
        { title: 'SEO关键字', dataIndex: 'keywords', width: 200, ellipsis: true },
        { title: 'SEO描述', dataIndex: 'description', width: 200, ellipsis: true },*/
        { title: t('浏览数'), dataIndex: 'views', width: 60, ellipsis: true  },
        { title: t('权重'), dataIndex: 'weight', width: 60, ellipsis: true  },
        { title: t('添加日期'), dataIndex: 'fmt_created_at', width: 120 },
        /*{ title: 'Update', dataIndex: 'fmt_updated_at', width: 160 },*/
        {
            title: t('状态'),
            dataIndex: 'status',
            width: 60,
            render: (text, record, index) => (
                <div>
                    {
                        record.status === 1 ? (
                            <Tag color='green'>
                                {t('启用')}
                            </Tag>
                        ) : (
                            <Tag color='red'>
                                {t('禁用')}
                            </Tag>
                        )
                    }
                </div>
            )
        },
        {
            title: '',
            dataIndex: 'operate',
            width: 180,
            render: (text, record, index) => {
                return (
                    <Space wrap>
                        <a href={`/doc/${encodeURIComponent(record.slug)}`} target='_blank'>
                            <Button theme='light' type='tertiary' style={{ marginBottom: 1 }}>
                                {t('查看')}
                            </Button>
                        </a>

                        <Popconfirm
                            title={t('确定是否要删除此文章？')}
                            content={t('此修改将不可逆')}
                            okType={'danger'}
                            position={'left'}
                            onConfirm={() => {
                                onDelete(record.id)
                            }}
                        >
                            <Button theme='light' type='danger' style={{ marginBottom: 1 }}>
                                {t('删除')}
                            </Button>
                        </Popconfirm>

                        <Button
                            theme='light'
                            type='tertiary'
                            style={{ marginBottom: 1 }}
                            onClick={() => {
                                setEditingDoc(record);
                                setShowEdit(true);
                            }}
                        >
                            {t('编辑')}
                        </Button>
                    </Space>
                );
            },
        }
    ];



    const { id,title } = inputs

    const formatDocList = useMemo(() => {
        return docList.map((item) => {
            item.fmt_created_at = formatDateTime(item.created_at, 'YYYY-MM-DD');
            item.fmt_updated_at = formatDateTime(item.updated_at, 'YYYY-MM-DD');
            return item
        })
    }, [docList])

    const handleInputChange = (value, name) => {
        setInputs(inputs => ({ ...inputs, [name]: value }));
    };

    const loadDocs = async (startIdx, pageSize) => {
        setLoading(true);
        const query = {
            p: startIdx,
            page_size: pageSize,
            ...inputs
        }

        let url = `/api/doc/?${new URLSearchParams(query).toString()}`;
        url = encodeURI(url);
        try {
            const res = await API.get(url);
            const { success, message, data } = res.data;
            if (success) {
                setDocList(data.items)
                setActivePage(data.page);
                setPageSize(data.page_size);
                setDocCount(data.total);
            } else {
                showError(message);
            }
        } catch (error) {
            showError(error);
        } finally {
            setLoading(false);

        }
    };

    const handlePageChange = (page) => {
        // 设置页数
        setActivePage(page);
        // 请求数据
        loadDocs(page, pageSize)
    };

    const handlePageSizeChange = async (size) => {
        localStorage.setItem('page-size', size + '');
        setPageSize(size);
        setActivePage(1);
        await loadDocs(activePage, size);

    };

    const refresh = async () => {
        setActivePage(1);
        loadDocs(1, pageSize);
    };

    useEffect(() => {
        const localPageSize =
            parseInt(localStorage.getItem('page-size')) || ITEMS_PER_PAGE;
        setPageSize(localPageSize);
        loadDocs(activePage, localPageSize)
    }, []);

    async function onDelete(id) {
        setLoading(true);
        let res = await API.delete(`/api/doc/${id}`)
        const { success, message } = res.data;
        if (success) {
            showSuccess(t('操作成功完成！'));
            refresh()
        } else {
            showError(message);
        }
        setLoading(false);
    }
    return (
        <>
            <Layout>
                <div className='flex justify-between items-center'>
                    <Form layout='horizontal' style={{ marginTop: 10 }}>
                        <>
                            <Form.Input
                                field='id'
                                label={t('ID')}
                                value={id}
                                placeholder={t('ID')}
                                name='title'
                                onChange={(value) => handleInputChange(value, 'id')}
                            />
                            <Form.Input
                                field='title'
                                label={t('标题')}
                                value={title}
                                placeholder={t('标题')}
                                name='title'
                                onChange={(value) => handleInputChange(value, 'title')}
                            />

                            <Button
                                label={t('查询')}
                                type='primary'
                                htmlType='submit'
                                className='btn-margin-right'
                                onClick={refresh}
                                loading={loading}
                                style={{ marginTop: 24 }}
                            >
                                {t('查询')}
                            </Button>
                        </>
                    </Form>
                    <Button label={t('添加')} type='primary' onClick={() => setShowEdit(true)}>
                        {t('添加')}
                    </Button>
                </div>


                <EditDoc
                    refresh={refresh}
                    editingDoc={editingDoc}
                    visible={showEdit}
                    handleClose={() => {
                        setEditingDoc({
                            id: undefined,
                        });
                        setShowEdit(false)
                    }}
                ></EditDoc>

                <Table
                    style={{ marginTop: 5 }}
                    columns={columns}
                    expandRowByClick={true}
                    dataSource={formatDocList}
                    rowKey="key"
                    pagination={{
                        formatPageText: (page) =>
                            t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                                start: page.currentStart,
                                end: page.currentEnd,
                                total: docCount
                            }),
                        currentPage: activePage,
                        pageSize: pageSize,
                        total: docCount,
                        pageSizeOpts: [10, 20, 50, 100],
                        showSizeChanger: true,
                        onPageSizeChange: (size) => {
                            handlePageSizeChange(size);
                        },
                        onPageChange: handlePageChange,
                    }}
                />
            </Layout>
        </>
    );
};

export default DocsTable;
