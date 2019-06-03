var	main= new Vue({
    el: '#main',
    data: {
        dialogUpdateVisible:false,
        dialogSaveVisible:false,
        tableData:[],
		ipUrl:'http://127.0.0.1:1803/getip',
        tableSearchModel: {
            ip: ''
        },
        totalElements:0,
        
        value4: '',
        formLabelWidth: '120px'
    },
    mounted: function () {
        this.refreshTable(this.ipUrl);
    },
    methods: {

        //查询按钮
        searchClick: function () {
            this.refreshTable(this.ipUrl);
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
                    // main.tableData.push(data)
                    main.tableData = data
                    // main.totalElements=parseInt(data.totalElements);
                },
                error:function(err){
                    console.log(err)
                }
            });
      }
    }
});