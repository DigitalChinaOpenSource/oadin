import { useState } from "react";
export function useViewModel() {
	const [modelSourceVal, setModelSourceVal] = useState<string>('local');
	const [modelPathVal, setModelPathVal] = useState<string>('');
	const [modelNums, setModelNums] = useState<number>(0);
	const [modelSearchVal, setModelSearchVal] = useState<string>('');

	const onModelSourceChange = (val: string) => {
		setModelSourceVal(val);
	}

	const onModelPathChange = (val: string) => {
		setModelPathVal(val);
	}

	const onModelTitleSearch = (val: string) => {
		setModelSearchVal(val);
	}

	return {
		modelSourceVal,
		onModelSourceChange,
		modelPathVal,
		onModelPathChange,
		modelNums,
		modelSearchVal,
		onModelTitleSearch
	}
}