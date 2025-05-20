function dayjs() {
    try {
        // YYYY-MM-DD
        const today = new Date();

        // 检查日期是否有效
        if (isNaN(today.getTime())) {
            console.error('Invalid date created');
            return null;
        }

        const year = today.getFullYear();
        const month = String(today.getMonth() + 1).padStart(2, '0');
        const day = String(today.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    } catch (error) {
        console.error('Error in dayjs function:', error);
        return null;
    }
}

export default dayjs;
