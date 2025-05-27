import styles from './index.module.scss';
import { Modal, Pagination } from 'antd';
import { useViewModel } from './view-model';
import modelPng from '@/assets/modelLogo.png';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import { IServiceProviderDataItem } from '../types';

export interface IServiceProviderDetailProps {
  selectedRow: IServiceProviderDataItem;
  onCancel: () => void;
}

export default function ServiceProviderDetail(props: IServiceProviderDetailProps) {
  const vm = useViewModel(props);

  return (
    <Modal
      centered
      open
      width={860}
      footer={null}
      title={<div className={styles.modalTitle}>查看服务提供商详情</div>}
      onCancel={vm.onCancel}
      okText="确认"
    >
      {vm.providerDetailLoading ? (
        <div className={styles.loading}>
          <img
            src={realLoadingSvg}
            alt="loading"
          />
        </div>
      ) : (
        <>
          <div className={styles.infoBlock}>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>服务提供商名称:</span>
              {vm.providerDetail.provider_name}
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>服务提供商厂商名称:</span>
              {vm.providerDetail.flavor}
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>服务来源:</span>
              {vm.providerDetail.service_source === 'remote' ? '云端' : '本地'}
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>服务名称:</span>
              {vm.providerDetail.service_name}
            </div>
            <div className={styles.infoItem}>
              <span className={styles.infoLabel}>服务提供商状态:</span>
              <div className={vm.providerDetail.status === 1 ? styles.readyStatus : styles.disabledStatus}>
                <div className={styles.dot}></div>
                {vm.providerDetail.status === 1 ? '可用' : '禁用'}
              </div>
            </div>
          </div>

          {vm.providerDetail?.support_model_list?.length > 0 && (
            <>
              <div className={styles.pageBlock}>
                <div className={styles.modelTitle}>支持的模型列表</div>
                <Pagination
                  current={vm.providerDetail?.page || 1}
                  showSizeChanger={false}
                  pageSize={vm.providerDetail?.page_size || 5}
                  total={vm.providerDetail?.total_count || 0}
                  onChange={vm.handlePageChange}
                  show-less-items
                />
              </div>
              <div className={styles.modelList}>
                {(vm.providerDetail?.support_model_list || []).map((model, index) => (
                  <div
                    className={styles.modelItem}
                    key={index}
                  >
                    <div className={styles.modelLeft}>
                      <img
                        src={modelPng}
                        alt="modelLogo"
                      />
                      <span className={styles.modelBaseInfo}>
                        <span className={styles.modelLabel}>模型名称: </span>
                        {model.name}
                      </span>
                      {model.class.map((tag, tagIndex) => (
                        <span
                          key={tagIndex}
                          className={styles.tagItem}
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                    <div className={styles.modelRight}>
                      {model.flavor}
                      <div className={styles.line}></div>
                      {Boolean(model.params_size) && (
                        <>
                          上下文长度: {model.params_size}k<div className={styles.line}></div>
                        </>
                      )}

                      {model.is_downloaded ? '已下载' : '未下载'}
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

          <div className={styles.timeInfo}>
            <span>创建时间: {vm.formateIsoTime(vm.providerDetail?.created_at)}</span>
            <span>更新时间: {vm.formateIsoTime(vm.providerDetail?.updated_at)}</span>
          </div>
        </>
      )}
    </Modal>
  );
}
