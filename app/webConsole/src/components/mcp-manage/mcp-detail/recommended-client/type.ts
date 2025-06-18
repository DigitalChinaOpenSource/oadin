export type CardItemType = {
  icon: string;
  name: string;
  description: string;
  id: string | number;
  linkCommand?: string;
};

export type CardItemProps = {
  clientData: CardItemType;
  handleClick: (id: string | number, data: CardItemType) => void;
};
