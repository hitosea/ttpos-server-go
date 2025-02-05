import router from './index.js'
// let  modules = import.meta.glob('../views/*/*/*.vue')
let modules = import.meta.glob(['../views/*.vue', '../views/*/*.vue', '../views/*/*/*.vue', '../views/*/*/*/*.vue',
	'../views/*/*/*/*/*.vue'
])

let count = 0;
const dealWithRoute = async (data, parent = 'Home') => {
	for (let item of data) {
		count = count + 1;
		item.component = modules[`../views${item.path}.vue`];
		//
		item.path = '/:catchAll(.*)' + item.path
		item.redirect_name && (item.redirect_name = '/:catchAll(.*)' + item.redirect_name)
		//
        router.addRoute(parent, item)
		if (item.children && item.children.length > 0) {
			dealWithRoute(item.children)
		}
	}
};

export default dealWithRoute;
