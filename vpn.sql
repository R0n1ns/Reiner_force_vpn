PGDMP      /    
            |            vpn    16.2    16.2                 0    0    ENCODING    ENCODING        SET client_encoding = 'UTF8';
                      false                       0    0 
   STDSTRINGS 
   STDSTRINGS     (   SET standard_conforming_strings = 'on';
                      false                       0    0 
   SEARCHPATH 
   SEARCHPATH     8   SELECT pg_catalog.set_config('search_path', '', false);
                      false                       1262    16519    vpn    DATABASE     w   CREATE DATABASE vpn WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'Russian_Russia.1251';
    DROP DATABASE vpn;
                postgres    false            �            1259    34130    products    TABLE     �  CREATE TABLE public.products (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    now_price bigint NOT NULL,
    last_price bigint DEFAULT 0,
    is_on_sale boolean DEFAULT false,
    is_term boolean DEFAULT false,
    term bigint DEFAULT 0,
    is_traffic boolean DEFAULT false,
    traffic bigint DEFAULT 0
);
    DROP TABLE public.products;
       public         heap    postgres    false            �            1259    34129    products_id_seq    SEQUENCE     x   CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 &   DROP SEQUENCE public.products_id_seq;
       public          postgres    false    218                       0    0    products_id_seq    SEQUENCE OWNED BY     C   ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;
          public          postgres    false    217            �            1259    42664    sales    TABLE     �  CREATE TABLE public.sales (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    userid bigint NOT NULL,
    peer text NOT NULL,
    config text NOT NULL,
    productid bigint NOT NULL,
    is_frozen boolean DEFAULT false,
    expiration_froze_date bigint,
    expiration_date timestamp with time zone,
    remaining_traffic numeric
);
    DROP TABLE public.sales;
       public         heap    postgres    false            �            1259    42663    sales_id_seq    SEQUENCE     u   CREATE SEQUENCE public.sales_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 #   DROP SEQUENCE public.sales_id_seq;
       public          postgres    false    220                       0    0    sales_id_seq    SEQUENCE OWNED BY     =   ALTER SEQUENCE public.sales_id_seq OWNED BY public.sales.id;
          public          postgres    false    219            �            1259    34075    users    TABLE     }  CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tgid bigint NOT NULL,
    mail text DEFAULT 'null'::text,
    password text DEFAULT 'null'::text,
    user_name text NOT NULL,
    isblocked boolean DEFAULT false,
    is_admin boolean DEFAULT false
);
    DROP TABLE public.users;
       public         heap    postgres    false            �            1259    34074    users_id_seq    SEQUENCE     u   CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
 #   DROP SEQUENCE public.users_id_seq;
       public          postgres    false    216                       0    0    users_id_seq    SEQUENCE OWNED BY     =   ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
          public          postgres    false    215            _           2604    34133    products id    DEFAULT     j   ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);
 :   ALTER TABLE public.products ALTER COLUMN id DROP DEFAULT;
       public          postgres    false    218    217    218            f           2604    42667    sales id    DEFAULT     d   ALTER TABLE ONLY public.sales ALTER COLUMN id SET DEFAULT nextval('public.sales_id_seq'::regclass);
 7   ALTER TABLE public.sales ALTER COLUMN id DROP DEFAULT;
       public          postgres    false    220    219    220            Z           2604    34078    users id    DEFAULT     d   ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
 7   ALTER TABLE public.users ALTER COLUMN id DROP DEFAULT;
       public          postgres    false    216    215    216            	          0    34130    products 
   TABLE DATA           �   COPY public.products (id, created_at, updated_at, deleted_at, name, now_price, last_price, is_on_sale, is_term, term, is_traffic, traffic) FROM stdin;
    public          postgres    false    218   U%                 0    42664    sales 
   TABLE DATA           �   COPY public.sales (id, created_at, updated_at, deleted_at, userid, peer, config, productid, is_frozen, expiration_froze_date, expiration_date, remaining_traffic) FROM stdin;
    public          postgres    false    220   �%                 0    34075    users 
   TABLE DATA           }   COPY public.users (id, created_at, updated_at, deleted_at, tgid, mail, password, user_name, isblocked, is_admin) FROM stdin;
    public          postgres    false    216   	&                  0    0    products_id_seq    SEQUENCE SET     >   SELECT pg_catalog.setval('public.products_id_seq', 44, true);
          public          postgres    false    217                       0    0    sales_id_seq    SEQUENCE SET     :   SELECT pg_catalog.setval('public.sales_id_seq', 1, true);
          public          postgres    false    219                       0    0    users_id_seq    SEQUENCE SET     :   SELECT pg_catalog.setval('public.users_id_seq', 8, true);
          public          postgres    false    215            q           2606    34143    products products_pkey 
   CONSTRAINT     T   ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);
 @   ALTER TABLE ONLY public.products DROP CONSTRAINT products_pkey;
       public            postgres    false    218            t           2606    42672    sales uni_sales_id 
   CONSTRAINT     P   ALTER TABLE ONLY public.sales
    ADD CONSTRAINT uni_sales_id PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.sales DROP CONSTRAINT uni_sales_id;
       public            postgres    false    220            j           2606    34086    users uni_users_id 
   CONSTRAINT     P   ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_id PRIMARY KEY (id);
 <   ALTER TABLE ONLY public.users DROP CONSTRAINT uni_users_id;
       public            postgres    false    216            l           2606    34088    users uni_users_tgid 
   CONSTRAINT     O   ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_tgid UNIQUE (tgid);
 >   ALTER TABLE ONLY public.users DROP CONSTRAINT uni_users_tgid;
       public            postgres    false    216            n           2606    34090    users uni_users_user_name 
   CONSTRAINT     Y   ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_user_name UNIQUE (user_name);
 C   ALTER TABLE ONLY public.users DROP CONSTRAINT uni_users_user_name;
       public            postgres    false    216            o           1259    34144    idx_products_deleted_at    INDEX     R   CREATE INDEX idx_products_deleted_at ON public.products USING btree (deleted_at);
 +   DROP INDEX public.idx_products_deleted_at;
       public            postgres    false    218            r           1259    42683    idx_sales_deleted_at    INDEX     L   CREATE INDEX idx_sales_deleted_at ON public.sales USING btree (deleted_at);
 (   DROP INDEX public.idx_sales_deleted_at;
       public            postgres    false    220            h           1259    34091    idx_users_deleted_at    INDEX     L   CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);
 (   DROP INDEX public.idx_users_deleted_at;
       public            postgres    false    216            u           2606    42678    sales fk_sales_product    FK CONSTRAINT     �   ALTER TABLE ONLY public.sales
    ADD CONSTRAINT fk_sales_product FOREIGN KEY (productid) REFERENCES public.products(id) ON UPDATE CASCADE ON DELETE CASCADE;
 @   ALTER TABLE ONLY public.sales DROP CONSTRAINT fk_sales_product;
       public          postgres    false    220    4721    218            v           2606    42673    sales fk_users_purchases    FK CONSTRAINT     v   ALTER TABLE ONLY public.sales
    ADD CONSTRAINT fk_users_purchases FOREIGN KEY (userid) REFERENCES public.users(id);
 B   ALTER TABLE ONLY public.sales DROP CONSTRAINT fk_users_purchases;
       public          postgres    false    4714    220    216            	   �   x���K
�0е|��C�4�7��	�m/Pߟ�ih��P0h1�CV$0�(27)���C+�:��I����ل���Ԭ`�bW�����|�������g5iN��=��z�M�������)���1�f#��b�ι�yF(            x������ � �         \   x�3��"#CcCΒ������0B�q�qYp����(Y�X��Y��j��nadbjjj``6��C�%1z\\\ �!"�     