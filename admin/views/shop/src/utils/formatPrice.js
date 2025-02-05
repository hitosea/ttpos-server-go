// 这个是有逗号的 返回类似 10,00,000
const formatprice = (price) => {
    if (isNaN(price)) {
        return 'Invalid Price';
    }

    // 将价格转换为数字类型
    const num = parseFloat(price);
    // 判断小数部分是否大于0
    if (num % 1 > 0) {
        return Number(num.toFixed(2).replace(/(\.0*|0+)$/, '')).toLocaleString('en-US'); // 保留小数，去掉末尾的0
    }
    else {
        return Number(num.toFixed(2)).toLocaleString('en-US'); // 不带小数部分
    }
};
export const formatPrice = formatprice;

// 这个是没有逗号的 返回类似 1000000
const pricetwo = (price) => {
    if (isNaN(price)) {
        return 'Invalid Price';
    }

    // 将价格转换为数字类型
    const num = parseFloat(price);

    // 判断小数部分是否大于0
    if (num % 1 > 0) {
        return Number(num.toFixed(2).replace(/(\.0*|0+)$/, '')); // 保留小数，去掉末尾的0
    }
    else {
        return Number(num.toFixed(2)); // 不带小数部分
    }
};
export const priceTwo = pricetwo;

const decimalPointFour = (price) => {
    if (isNaN(price)) {
        return 'Invalid Price';
    }

    const num = parseFloat(price);
    const decimal = num.toString().split('.')[1] || ''; // 获取小数部分，若无小数部分则为空字符串

    let formattedNum;
    if (decimal === '') {
        formattedNum = num.toFixed(0); // 如果小数部分为空，则只保留整数部分
    } else {
        const decimalsToKeep = Math.min(decimal.length, 4);
        formattedNum = num.toFixed(decimalsToKeep).replace(/(\.0*|0+)$/, ''); // 最多保留4位小数，去掉多余的0
    }

    return Number(formattedNum);
};

export const DecimalPointFour = decimalPointFour;

const priceCalculation = (price) => {
    if (isNaN(price)) {
        return 'Invalid Price';
    }

    // 将价格转换为数字类型
    const num = parseFloat(price);

    // 判断小数部分是否大于0
    if (num % 1 > 0) {
        return Number(Number(Math.round(num * 100) / 100).toFixed(2).replace(/(\.0*|0+)$/, '')).toLocaleString('en-US'); // 四舍五入保留2小数，去掉末尾的0
    }
    else {
        return Number(num.toFixed(2)).toLocaleString('en-US'); // 不带小数部分
    }
};

export const PriceCalculation = priceCalculation;