import { useRef, useImperativeHandle, forwardRef } from 'react';
import '@toast-ui/editor/dist/i18n/zh-cn';
import '@toast-ui/editor/dist/toastui-editor.css';
import { Editor } from '@toast-ui/react-editor';
import { API, showError } from '../helpers';
export default forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    // 设置内容
    setContent: (content) => {
      const editorInstance = editorRef.current.getInstance();
      editorInstance.setHTML(content);
    },
  }));
  const editorRef = useRef();
  // 修改内容
  function changeContent() {
    const editorInstance = editorRef.current.getInstance();
    const content = editorInstance.getHTML();
    props.changeContent(content);
  }
  // 图片上传
  async function onAddImageBlobHook(file, callback) {
    const editorInstance = editorRef.current.getInstance();
    const formData = new FormData();
    formData.append('file', file);
    const res = await API.post('/api/file/upload', formData);
    if (res.data.success) {
      const imageUrl = res.data.data.path;
      callback(imageUrl, '');
      editorInstance.focus();
    } else {
      showError(res.data.message);
    }
  }
  return (
    <Editor
      ref={editorRef}
      language='zh-CN'
      usageStatistics={false}
      initialValue=' '
      onChange={changeContent}
      hooks={{ addImageBlobHook: onAddImageBlobHook }}
    />
  );
});
