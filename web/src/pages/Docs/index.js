import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { Empty, Pagination, Spin } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import "./docs.css";
const Docs = () => {
  const { t } = useTranslation();
  const [docList, setDocList] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [params, setParams] = useState({
    p: 1,
    keyword: "",
  });
  const [pagination, setPagination] = useState({
    page_size: 0,
    total: 0,
  });

  const loadDocList = async (page = 1, keyword = "") => {
    setIsLoading(true);
    try {
      const res = await API.get(`/api/doc/list?p=${page}&keyword=${keyword}`);
      const { success, message, data } = res.data;
      if (success) {
        const { items, page_size, total } = data
        setDocList(items);
        setPagination({
          page_size,
          total,
        })
      } else {
        showError(message);
      }
    } catch (error) {
    } finally {
      setIsLoading(false);
    }
  };
  function onChange(p) {
    setParams({
      ...params,
      p,
    })
    loadDocList(p, params.keyword)
  }
  function changeKeyword(e) {
    console.log(e)
    setParams({
      ...params,
      keyword: e.target.value,
    })
  }
  useEffect(() => {
    loadDocList();
  }, []);

  return (
    <div className="mt-[64px]">
      <div id="burncloud-section">
        <div className="burncloud-container-docs">
          <div className="container">
            <section className="search-section">
              <div className="search-container">
                <input type="text" className="search-input" id="searchInput" name="keywords" placeholder="Search documents..." onChange={changeKeyword} defaultValue={params.keyword} />
                <button className="search-button" type="submit" onClick={() => {
                  loadDocList(params.p, params.keyword)
                }}>🔍</button>
              </div>
            </section>
          </div>

          <div className="out_container_product-list faq-section" id="faqAncher">
            <div className="container">
              {isLoading ? (
                <div class="flex justify-center mb-4"><Spin /></div>
              ) : null}
              {docList.length > 0 ? (
                <>
                  {docList.map(item => (
                    <div className="faq-item" key={item.id}>
                      <div className="faq-question">
                        <a href={`/doc/${encodeURIComponent(item.slug)}`}>
                          {item.title}
                        </a>
                      </div>
                      <div className="faq-answer">
                        {item.summary}
                      </div>
                    </div>
                  ))}
                  <Pagination total={pagination.total} pageSize={pagination.page_size} style={{ marginBottom: 12 }} onChange={onChange}></Pagination>
                </>
              ) : <Empty title={t('无记录')} />}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Docs;
