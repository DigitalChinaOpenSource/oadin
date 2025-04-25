import { useState } from "react";
export function useViewModel() {
	const [modelSourceVal, setModelSourceVal] = useState<string>('local');
	const [modelSearchVal, setModelSearchVal] = useState<string>('');

	const onModelSourceChange = (val: string) => {
		setModelSourceVal(val);
	}

	const onModelSearch = (val: string) => {
		setModelSearchVal(val);
	}

	return {
		modelSourceVal,
		onModelSourceChange,
		modelSearchVal,
		onModelSearch
	}
}