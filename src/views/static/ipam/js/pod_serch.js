var	index= new Vue({
    el: '#index',
    data: {
        dialogUpdateVisible:false,
        dialogSaveVisible:false,
        tableData:[],
		podUrl:'http://127.0.0.1:1803/getpodname',
        tableSearchModel: {
			namespace:'',
            podname:'',
            ip: '',
            status:1,
            size: 10,
            page: 1
        },
        statusSet:[{
            value: null,
            label: ''
        },{
            value: 1,
            label: '使用中'
        },{
            value: 0,
            label: '已释放'
        }],
        totalElements:0,
        value4: '',
        formLabelWidth: '120px'
    },
    mounted: function () {
        this.refreshTable(this.podUrl);
    },
    methods: {

        //查询按钮
        searchClick1: function () {
            this.refreshTable(this.podUrl);
        },
        //表的优先级控制
        tableRowClassName :function(row, rowIndex) {
          if (row.row.priority == 1) {
              return 'a1-row';
          } else if (row.row.priority == 2) {
              return 'a2-row';
          }else if (row.row.priority == 3) {
              return 'a3-row';
          }
          return 'a4-row';
      },
        //页面事件数量设置
        handleSizeChange:function(val) {
            this.tableSearchModel.pageSize=val;
            this.refreshTable(this.podUrl);
        },
        //跳转页面
        handleCurrentChange:function(val) {
            this.tableSearchModel.page=val;
            this.refreshTable(this.podUrl);
        },
        //刷新表数据
        refreshTable :function(url) {
            $.ajax({
                url: url,
                type: 'GET',
                dataType: 'json',    //第1处
                contentType:'application/json;charset=UTF-8',
                data:
                    this.tableSearchModel
                ,
                success:function(data){
                    for(var i=0;i<data.Podinfoes.length;i++) {
                        if(data.Podinfoes[i].Status==1){
                            data.Podinfoes[i].Status="使用中";
                        }else{
                            data.Podinfoes[i].Status="已释放";
                        }
                    }
                    // index.tableData.push(data)
                    index.tableData=data.Podinfoes
                    index.totalElements=parseInt(data.Count);
                },
                error:function(err){
					console.log(err)
                }
            });
      }
    }
});