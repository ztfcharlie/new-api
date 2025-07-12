import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { Empty } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import "./doc.css";
import { useParams } from 'react-router-dom';
const DocInfo = () => {
  const { t } = useTranslation();
  const [docDetail, setDocDetail] = useState({});

  const loadDocInfo = async (slug) => {
    try {
      const res = await API.get(`/api/doc/detail?slug=${slug}`);
      const { success, message, data } = res.data;
      if (success) {
        setDocDetail(data);
      } else {
        showError(message);
      }
    } catch (error) {
    } finally {
    }
  };

  const { slug } = useParams();
  useEffect(() => {
    loadDocInfo(slug);
  }, []);

  return (
    <div className="mt-[64px]">
      <div className="burncloud-container-docs-detail">
        <div className="container">
          <div className="breadcrumb">
            <a href="/">Home</a> / <a href="/docs">Docs</a> / {docDetail.title}
          </div>
          <div className="divider" />
          {docDetail.content ? (
            <>
              <div className="article-title">
                {docDetail.title}
              </div>
              <div className="article-meta">By Burncloud &nbsp; | &nbsp; {docDetail.created_at}</div>
              <div className="article-summary">
                {docDetail.summary}
              </div>
              <div className="article-content">
                <div
                  style={{ fontSize: 'larger' }}
                  dangerouslySetInnerHTML={{ __html: docDetail.content }}
                ></div>
              </div>
            </>
          ) : <Empty title={t('无记录')} />}
        </div>
      </div>

    </div>
  );
};

export default DocInfo;
