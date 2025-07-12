import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';

const Faq = () => {
    const { t,i18n } = useTranslation();
    const [faq, setFaq] = useState('');
  
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
    };
  
    useEffect(() => {
        displayFaq().then();
    }, []);
  
    useEffect(() => {
        const handleLanguageChanged = (lng) => {
            displayFaq();
        };

        i18n.on('languageChanged', handleLanguageChanged);
        return () => {
            i18n.off('languageChanged', handleLanguageChanged);
        };
    }, [i18n]);
    return (
        <div className="mt-[64px]">
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
        </div>
        
    );
};
  
export default Faq;