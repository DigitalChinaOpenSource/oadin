import styles from './index.module.scss';
import {Col,Row} from 'antd';
import ModelCard from './model-card';

export default function ModelCardList() {
  return (
    <div className={styles.modelCardList}>
      <Row gutter={[16,16]}>
        {[1,2,3,4,5,6,7,8].map((item, index) => {
          return (
            <Col xs={24} sm={24} md={24} lg={12} xl={12} span={4} key={index}>
              <ModelCard />
            </Col>
          )
        })}
      </Row>
    </div>
  );
};