#!/bin/sh
# 启动前的准备工作
echo "Starting application..."

echo "input args:""$@"
POD_NAME=${POD_NAME:-"my-go-app2-1111-904-0-8wtsr"}  # 这里可以模拟 POD_NAME 的值

#echo $POD_NAME
# 使用 grep 和 sed 提取数字后面的部分
pod_id=$(echo "$POD_NAME" | awk -F'-' '{print $(NF-1)}')
# 你可以在这里添加一些初始化代码
# 例如检查环境变量、创建目录、等待服务等等

#echo "$@"
# 运行传入的命令

# 初始化变量
total_robot_num=0
pod_num=""
#pod_id=""
account_on=false
account_filename=""
account_prefix=""
account_idx=""
log_on=""
log_level=""
consecutively=""
#report_id=""
cases=""
modified_args=""
#echo "$#"
# 读取参数
#for arg in "$@"; do
#    echo $arg
#done
#
#modified_args="$modified_args --pod-id"
#modified_args="$modified_args $index

while [ "$#" -gt 0 ]; do #参数的个数
#    echo "主参数"$1;
    modified_args="$modified_args $1"
    case "$1" in
        --total-robot-num) total_robot_num="$2";
          modified_args="$modified_args $2"
          shift ;;
        --pod-num) pod_num="$2";
          modified_args="$modified_args $2"
          shift ;;
#        --pod-id) pod_id="$2";
#          modified_args="$modified_args $2"
#          shift ;;
        --account-on) account_on="$2";
          modified_args="$modified_args $2"
          shift;;
        --account-filename) account_filename="$2";
          modified_args="$modified_args $2"
          shift ;;
        --account-prefix) account_prefix="$2";
          modified_args="$modified_args $2"
          shift ;;
        --account-idx) account_idx="$2";
          quotient=$((total_robot_num / pod_num))
          remainder=$((total_robot_num % pod_num))
          pre_total_num=$((quotient * $pod_id))  #pod_id是n, 就说明前面有几个pod
          i=0
          while [ $i -lt $pod_id ]; do
            now_people=0
            if [ $i -lt $remainder ]; then  # 如果 i 小于余数，说明这个 pod 有多一个人
              pre_total_num=$(expr $pre_total_num + 1)
            else
              break  # 否则就跳出循环
            fi
            i=$(expr $i + 1)  # 遍历到下一个 pod
          done

#          for ((i=0; i<$pod_id; i=i+1)); do  #遍历0-pod_id之前的pod，看看哪些pod有多一个人
#            now_people=0
#            if [ $i -lt $remainder ]; then  #如果i小于余数，说明这个pod有多一个人
#              pre_total_num=$((pre_total_num + 1))
#            else
#              break #否则就跳出循环
#            fi
##            pre_total_num=$((pre_total_num + now_people))
#          done
          account_idx=$((pre_total_num + account_idx))
#          now_people=
#          if [ $pod_id -lt $remainder ]; then
#            now_people=$((quotient + 1))
#          else
#            now_people=$quotient
#          fi
#          account_idx=$((account_idx + now_people))
          modified_args="$modified_args $account_idx"

#          ((account_idx = account_idx + now_people))
          shift ;;
        --log-on) log_on="$2";
          modified_args="$modified_args $2"
          shift ;;
        --log-level) log_level="$2";
          modified_args="$modified_args $2"
          shift ;;
        --consecutively) consecutively="$2";
          modified_args="$modified_args $2"
          shift;;
        --case)#--case 是 $1
              case_name="$2"
              modified_args="$modified_args $2"
              case_robot_num=""
              robot_num=""
              pod_robot_num=""
              ramp_up_period=""
              duration=""
              loop_num=""
              timeout=""
  #            echo $case_name
              while [ "$#" -gt 2 ]; do
                  modified_args="$modified_args $3"
                  case "$3" in
                      --case-robot-num) case_robot_num="$4";
                        modified_args="$modified_args $4"

                        quotient=$((case_robot_num / pod_num))
                        remainder=$((case_robot_num % pod_num))
                        now_people=0
                        if [ $pod_id -lt $remainder ]; then
                          now_people=$((quotient + 1))
                        else
                          now_people=$quotient
                        fi
                        pod_robot_num=$now_people
                        #重新赋值，每个pod的人数根据pod_id重新计算
                        modified_args="$modified_args --pod-robot-num"
                        modified_args="$modified_args $now_people"
                        shift ;;
                      --ramp-up-period) ramp_up_period="$4";
                        modified_args="$modified_args $4"
                        shift ;;
                      --duration) duration="$4";
                        modified_args="$modified_args $4"
                        shift ;;
                      --loop-num) loop_num="$4";
                        modified_args="$modified_args $4"
                        shift ;;
                      --timeout) timeout="$4";
                        modified_args="$modified_args $4"
                        shift ;;
                      --case)  break ;;
                      *) value="$4"; #动态参数
                        modified_args="$modified_args $value"
                        shift ;;
                  esac
                  shift  #消除子项参数
              done

              shift #消除case_name
            # 将案例存入字符串
              cases="$cases\nName: $case_name, Total Robot Num: $total_robot_num, pod_robot_num: $pod_robot_num, Robot Num: $robot_num, Ramp-up Period: $ramp_up_period, Duration: $duration, Loop Num: $loop_num, Timeout: $timeout"
              ;;

        *)
#           modified_args="$modified_args $2"
#                    shift ;;
          echo "Unknown parameter: $1" ;;
    esac
    shift   #消除主项参数项，包括--case

done

#echo "修改后的参数"
#
#echo $modified_args
## 输出解析结果
#echo "解析结果:\n"
#echo "Total Robot Num: $total_robot_num"
#echo "Pod Num: $pod_num"
#echo "Pod ID: $pod_id"
#echo "Account On: $account_on"
#echo "Account Filename: $account_filename"
#echo "Account Prefix: $account_prefix"
#echo "Account Index: $account_idx"
#echo "Log On: $log_on"
#echo "Log Level: $log_level"
#echo "Consecutively: $consecutively"
#
#echo "\n解析Cases:"
#echo -e "\nCases:"
#printf "$cases"
#echo $cases

echo "trans args: "$modified_args
#echo "$@"
./stressTest --pod-id $pod_id  $modified_args
#最后关闭 vector
pkill -TERM -f '/root/vector/vector.*toml' || true

#read -p "按 Enter 键退出..."