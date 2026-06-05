using UnityEngine;
using System.Collections;
using System.Collections.Generic;
using System.Data;
using System.IO;
using System.Text;
using System.Reflection;
using System.Reflection.Emit;
using System;
using Excel;
using Newtonsoft.Json;
using System.Runtime.InteropServices.ComTypes;

//表格转Json文件

public class ExcelUtility
{
    /// <summary>
    /// 表格数据集合
    /// </summary>
    public DataSet mResultSet;
    public string excelFile;
    /// <summary>
    /// 构造函数
    /// </summary>
    /// <param name="excelFile">Excel file.</param>
    public ExcelUtility(string excelFile)
    {
        this.excelFile = excelFile;
        //FileStream mStream = File.Open(excelFile, FileMode.Open, FileAccess.Read);
        using FileStream mStream = File.Open(excelFile, FileMode.OpenOrCreate, FileAccess.Read);
        using IExcelDataReader excelReader = ExcelReaderFactory.CreateOpenXmlReader(mStream);
        mResultSet = excelReader.AsDataSet();

    }

    /// <summary>
    /// 转换为Json
    /// string  int float  double  bool
    /// </summary>
    /// <param name="JsonPath">Json文件路径</param>
    /// <param name="Header">表头行数</param>
    public string ConvertToJson()
    {
        //判断Excel文件中是否存在数据表
        if (mResultSet.Tables.Count < 1)
        {
            //一个文件一个表格
            Debug.Log(excelFile + "文件内出现多张表格");
            return null;
        }
        string json;

        //默认读取第一个数据表
        DataTable mSheet = mResultSet.Tables[0];

        string outname = mSheet.TableName;
        if (outname.IndexOf('#') >= 0 && outname.LastIndexOf('#') != outname.IndexOf('#'))
        {
            outname = outname.Substring(outname.IndexOf('#') + 1, outname.LastIndexOf('#') - outname.IndexOf('#') - 1);
        }
        else
        {
            Debug.Log("无法世界导出名 " + outname + "  请确定#书写正确!");
            return null;
        }

        //判断数据表内是否存在数据
        if (mSheet.Rows.Count < 1)
            return null;

        //读取数据表行数和列数
        int rowCount = mSheet.Rows.Count;
        int colCount = mSheet.Columns.Count;

        //准备一个列表存储整个表的数据
        List<Dictionary<string, object>> table = new List<Dictionary<string, object>>();

        //读取数据
        for (int i = 3; i < rowCount; i++)
        {
            //准备一个字典存储每一行的数据
            Dictionary<string, object> row = new Dictionary<string, object>();
            for (int j = 0; j < colCount; j++)
            {
                //读取第1行数据作为表头字段
                string field = mSheet.Rows[1][j].ToString();
                field = field.Trim();

                string typestring = mSheet.Rows[2][j].ToString();
                typestring = typestring.ToLower().Trim();

                string valuestr = mSheet.Rows[i][j].ToString();
                valuestr = valuestr.Trim();
                //Key-Value对应 按类型存放
                switch (typestring)
                {
                    case "int":
                        if (valuestr != "")
                        {
                            row[field] = Convert.ToInt32(valuestr);
                        }
                        else
                        {
                            row[field] = 0;
                        }
                        break;
                    case "float":
                        if (valuestr != "")
                        {
                            row[field] = float.Parse(valuestr);
                        }
                        else
                        {
                            row[field] = 0;
                        }
                        break;
                    case "double":
                        if (valuestr != "")
                        {
                            row[field] = Convert.ToDouble(valuestr);
                        }
                        else
                        {
                            row[field] = 0;
                        }

                        break;
                    case "bool":
                        if (valuestr == "0" || valuestr == "fasle" || valuestr == "")
                        {
                            row[field] = false;
                        }
                        else
                        {
                            row[field] = true;
                        }
                        break;
                    default:
                        row[field] = valuestr;
                        break;
                }
            }

            //添加到表数据中
            table.Add(row);
        }

        //生成Json字符串
        json = JsonConvert.SerializeObject(table, Newtonsoft.Json.Formatting.Indented);

        json = "{\n\"questions\":" + json + "\n}";
        return json;
        //写入文件
        //using (FileStream fileStream = new FileStream(JsonPath + "/" + outname + ".txt", FileMode.Create, FileAccess.Write))
        //{
        //    using (TextWriter textWriter = new StreamWriter(fileStream, encoding))
        //    {
        //        textWriter.Write(json);
        //        Debug.Log(json);
        //        Debug.Log(JsonPath+".json生成完成");
        //    }
        //}
    }


}

