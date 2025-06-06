import styles from './index.module.scss';
import favicon from '@/assets/favicon.png';
import { Button } from 'antd';
import { RightOutlined, ExportOutlined } from '@ant-design/icons';
import UpdateHistory from '@/components/settings/about-us/update-history';
import { useEffect, useState } from 'react';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore.ts';
import CopyrightNotice from '@/components/settings/about-us/copyright-notice';
import Feedback from '@/components/settings/about-us/feedback';
import { useAboutUsView } from '@/components/settings/about-us/view.module.ts';
import FetchLoading from '@/components/fetch-loading';
export default function AboutUs() {
  // 更新日志
  const [openHistory, setOpenHistory] = useState(false);
  // 版权声明
  const [openCopyright, setOpenCopyright] = useState(false);
  // 意见反馈
  const [openFeedback, setOpenFeedback] = useState(false);
  // 白泽服务状态
  const { checkByzeStatus, fetchByzeServerStatus } = useByzeServerCheckStore();
  // 关于我们详情
  const { aboutDetails, aboutUsLoading, fetchAboutUsDetail } = useAboutUsView();

  useEffect(() => {
    fetchAboutUsDetail();
  }, []);

  return (
    <div className={styles.aboutUs}>
      {aboutUsLoading ? (
        <FetchLoading />
      ) : (
        <div className={styles.mainContent}>
          <div className={styles.aboutUsTitle}>关于我们</div>
          <div className={styles.comCard}>
            <div className={styles.comItem}>
              <div className={styles.header}>
                <div className={styles.headerIcon}>
                  <img
                    src={aboutDetails?.logo || favicon}
                    alt=""
                  />
                </div>
                <div className={styles.headerContent}>
                  <div className={styles.headerTitle}>{aboutDetails?.name}</div>
                  <div className={styles.headerDesc}>{aboutDetails?.description}</div>
                </div>
              </div>
            </div>
            <div className={styles.comItem}>当前版本 &nbsp;&nbsp;&nbsp;&nbsp;{aboutDetails?.version}</div>
            <div className={styles.comItem}>
              <div className={styles.serviceStatus}>
                <div>服务状态</div>
                <div className={styles.statusName}>
                  {checkByzeStatus ? (
                    <>
                      <div className={styles.dot}></div>
                      <span className={styles.used}>启用</span>
                    </>
                  ) : (
                    <>
                      <div className={styles.noCheck}></div>
                      <span className={styles.stopped}>异常</span>
                    </>
                  )}
                </div>
              </div>
              <Button
                type="default"
                onClick={fetchByzeServerStatus}
              >
                重启
              </Button>
            </div>
          </div>
          <div className={styles.comCard}>
            {/*<div className={styles.comItem}>*/}
            {/*  <div>更新日志</div>*/}
            {/*  <RightOutlined*/}
            {/*    className={styles.comIcon}*/}
            {/*    onClick={() => setOpenHistory(true)}*/}
            {/*  />*/}
            {/*</div>*/}
            <div className={styles.comItem}>
              <div>版权声明</div>
              <RightOutlined
                className={styles.comIcon}
                onClick={() => setOpenCopyright(true)}
              />
            </div>
            <div className={styles.comItem}>
              <div>官方网站</div>
              <ExportOutlined
                className={styles.comIcon}
                onClick={() => aboutDetails?.officialWebsite && window.open(aboutDetails?.officialWebsite)}
              />
            </div>
            <div className={styles.comItem}>
              <div>意见反馈</div>
              <Button
                type={'default'}
                onClick={() => setOpenFeedback(true)}
              >
                反馈
              </Button>
            </div>
          </div>
        </div>
      )}
      <div className={styles.bottomDesc}>本产品基于AOG服务框架开发，遵循 Apache License2.0 协议开源</div>
      <UpdateHistory
        open={openHistory}
        onClose={() => setOpenHistory(false)}
      />
      <CopyrightNotice
        open={openCopyright}
        onClose={() => setOpenCopyright(false)}
        notice={aboutDetails?.copyright}
      />
      <Feedback
        open={openFeedback}
        onClose={() => setOpenFeedback(false)}
      />
    </div>
  );
}
