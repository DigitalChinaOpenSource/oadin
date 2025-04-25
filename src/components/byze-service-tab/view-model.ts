export function useViewModel() {
    const handleRefresh = () => {
        console.log('--- 状态刷新')
    }
    return {handleRefresh}
}