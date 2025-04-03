import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import { Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const Faq = () => {
    const { t } = useTranslation();
    const [faq, setFaq] = useState('');
    const [faqLoaded, setFaqLoaded] = useState(false);
  
    const displayFaq = async () => {
        setFaq(localStorage.getItem('faq') || '');
        const res = await API.get('/api/faq');
        const { success, message, data } = res.data;
        if (success) {
            let faqContent = data;
            if (!data.startsWith('https://')) {
                faqContent = marked.parse(data);
            }
            setFaq(faqContent);
            localStorage.setItem('faq', faqContent);
        } else {
            showError(message);
            setFaq(t('获取FAQ内容失败'));
        }
        setFaqLoaded(true);
    };
  
    useEffect(() => {
        displayFaq().then();
    }, []);
  
    return (
        <Layout>
            <Layout.Content>
                {faqLoaded && (
                    <>
                        {faq.startsWith('https://') ? (
                            <iframe
                                src={faq}
                                style={{ width: '100%', height: '100vh', border: 'none' }}
                            />
                        ) : (
                            <div
                                style={{ fontSize: 'larger' }}
                                dangerouslySetInnerHTML={{ __html: faq }}
                            ></div>
                        )}
                    </>
                )}
            </Layout.Content>
        </Layout>
    );
};
  
export default Faq;