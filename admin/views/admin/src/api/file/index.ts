import { $post } from '../request';

// 上传图片
// export interface UploadFileType {
//   iFile?: File; // 文件
// }
export function fetchUploadFile(data: FormData, headers: any = { 'Content-Type': 'multipart/form-data' }) {
  return $post('/file.Upload/image', data, { headers });
}
export function uploadFile(data: FormData, headers: any = { 'Content-Type': 'multipart/form-data' }) {
  return $post('/client.client/upload', data, { headers });
}
